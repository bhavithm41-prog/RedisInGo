package command

import (
	"testing"

	"github.com/bhavithm41-prog/gocachedb/internal/metrics"
	"github.com/bhavithm41-prog/gocachedb/internal/store"
)

// newTestHandler creates a Handler backed by a fresh, isolated
// Store and Metrics instance, using a temp-file path for the data
// file so tests never touch a real dump.rdb.
func newTestHandler(t *testing.T) *Handler {
	t.Helper()
	dataFile := t.TempDir() + "/test-dump.rdb"
	m := metrics.New()
	s := store.New(100, m)
	return New(s, dataFile, m)
}

func TestPing(t *testing.T) {
	h := newTestHandler(t)
	got := h.Execute("PING")
	want := "+PONG\r\n"
	if got != want {
		t.Fatalf("PING = %q, want %q", got, want)
	}
}

func TestSetAndGet(t *testing.T) {
	h := newTestHandler(t)

	got := h.Execute("SET name Bhavith")
	want := "+OK\r\n"
	if got != want {
		t.Fatalf("SET = %q, want %q", got, want)
	}

	got = h.Execute("GET name")
	want = "$7\r\nBhavith\r\n"
	if got != want {
		t.Fatalf("GET name = %q, want %q", got, want)
	}
}

func TestGetMissingKeyReturnsNullBulkString(t *testing.T) {
	h := newTestHandler(t)
	got := h.Execute("GET doesnotexist")
	want := "$-1\r\n"
	if got != want {
		t.Fatalf("GET missing key = %q, want %q", got, want)
	}
}

func TestUnknownCommand(t *testing.T) {
	h := newTestHandler(t)
	got := h.Execute("FOOBAR")
	want := "-ERR unknown command 'FOOBAR'\r\n"
	if got != want {
		t.Fatalf("unknown command = %q, want %q", got, want)
	}
}

func TestWrongNumberOfArguments(t *testing.T) {
	h := newTestHandler(t)
	got := h.Execute("SET onlyonearg")
	want := "-ERR wrong number of arguments for 'SET'\r\n"
	if got != want {
		t.Fatalf("SET with 1 arg = %q, want %q", got, want)
	}
}

func TestCaseInsensitiveCommands(t *testing.T) {
	h := newTestHandler(t)
	h.Execute("SET name Bhavith")

	got := h.Execute("get name")
	want := "$7\r\nBhavith\r\n"
	if got != want {
		t.Fatalf("lowercase get = %q, want %q", got, want)
	}
}

func TestDelAndExists(t *testing.T) {
	h := newTestHandler(t)
	h.Execute("SET key1 value1")

	got := h.Execute("EXISTS key1")
	if got != ":1\r\n" {
		t.Fatalf("EXISTS before DEL = %q, want :1\\r\\n", got)
	}

	got = h.Execute("DEL key1")
	if got != ":1\r\n" {
		t.Fatalf("DEL = %q, want :1\\r\\n", got)
	}

	got = h.Execute("EXISTS key1")
	if got != ":0\r\n" {
		t.Fatalf("EXISTS after DEL = %q, want :0\\r\\n", got)
	}
}

func TestListCommands(t *testing.T) {
	h := newTestHandler(t)

	got := h.Execute("RPUSH mylist a b c")
	want := ":3\r\n"
	if got != want {
		t.Fatalf("RPUSH = %q, want %q", got, want)
	}

	got = h.Execute("LRANGE mylist 0 -1")
	want = "*3\r\n$1\r\na\r\n$1\r\nb\r\n$1\r\nc\r\n"
	if got != want {
		t.Fatalf("LRANGE = %q, want %q", got, want)
	}
}

func TestWrongTypeError(t *testing.T) {
	h := newTestHandler(t)
	h.Execute("SET name Bhavith")

	got := h.Execute("LPUSH name x")
	want := "-ERR WRONGTYPE Operation against a key holding the wrong kind of value\r\n"
	if got != want {
		t.Fatalf("LPUSH on string key = %q, want %q", got, want)
	}
}

func TestExpireAndTTL(t *testing.T) {
	h := newTestHandler(t)
	h.Execute("SET name Bhavith")

	got := h.Execute("TTL name")
	if got != ":-1\r\n" {
		t.Fatalf("TTL with no expiry = %q, want :-1\\r\\n", got)
	}

	got = h.Execute("EXPIRE name 100")
	if got != ":1\r\n" {
		t.Fatalf("EXPIRE = %q, want :1\\r\\n", got)
	}
}

func TestInfoReturnsBulkString(t *testing.T) {
	h := newTestHandler(t)
	h.Execute("SET a 1")
	h.Execute("GET a")

	got := h.Execute("INFO")
	// We don't assert the exact byte count (it varies with
	// uptime's digit count), just that it's a properly-formed
	// RESP Bulk String starting with "$" and containing expected
	// metric field names.
	if len(got) == 0 || got[0] != '$' {
		t.Fatalf("INFO reply should be a RESP Bulk String, got %q", got)
	}
	if !contains(got, "total_get_commands:1") {
		t.Fatalf("INFO reply missing expected metric, got %q", got)
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) &&
		(func() bool {
			for i := 0; i+len(needle) <= len(haystack); i++ {
				if haystack[i:i+len(needle)] == needle {
					return true
				}
			}
			return false
		})()
}
