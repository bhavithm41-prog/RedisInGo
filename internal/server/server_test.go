package server

import (
	"bufio"
	"net"
	"testing"

	"github.com/bhavithm41-prog/gocachedb/internal/metrics"
	"github.com/bhavithm41-prog/gocachedb/internal/store"
)

// startTestServer creates a real Server backed by an isolated
// Store/Metrics pair, binds it to an OS-assigned free port (":0"),
// and starts serving in a background goroutine. Each call gets its
// own listener and port, so tests never collide with each other or
// with a real server you might be running manually.
func startTestServer(t *testing.T) string {
	t.Helper()

	dataFile := t.TempDir() + "/test-dump.rdb"
	m := metrics.New()
	s := store.New(100, m)
	srv := New("0", s, dataFile, m)

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("failed to create test listener: %v", err)
	}

	addr := listener.Addr().String()

	go func() {
		_ = srv.Serve(listener)
	}()

	return addr
}

func sendCommand(t *testing.T, addr, command string) string {
	t.Helper()

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("failed to connect to test server: %v", err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte(command + "\n"))
	if err != nil {
		t.Fatalf("failed to write command: %v", err)
	}

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}
	return line
}

func TestServerEndToEnd(t *testing.T) {
	addr := startTestServer(t)

	got := sendCommand(t, addr, "PING")
	want := "+PONG\r\n"
	if got != want {
		t.Fatalf("PING over real TCP = %q, want %q", got, want)
	}
}

func TestServerSetAndGetOverTCP(t *testing.T) {
	addr := startTestServer(t)

	got := sendCommand(t, addr, "SET name Bhavith")
	want := "+OK\r\n"
	if got != want {
		t.Fatalf("SET over real TCP = %q, want %q", got, want)
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	conn.Write([]byte("GET name\n"))
	reader := bufio.NewReader(conn)

	lengthLine, _ := reader.ReadString('\n')
	if lengthLine != "$7\r\n" {
		t.Fatalf("GET length line = %q, want %q", lengthLine, "$7\r\n")
	}
	valueLine, _ := reader.ReadString('\n')
	if valueLine != "Bhavith\r\n" {
		t.Fatalf("GET value line = %q, want %q", valueLine, "Bhavith\r\n")
	}
}

func TestServerConcurrentClients(t *testing.T) {
	addr := startTestServer(t)

	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func() {
			got := sendCommand(t, addr, "PING")
			done <- got == "+PONG\r\n"
		}()
	}

	for i := 0; i < 5; i++ {
		if ok := <-done; !ok {
			t.Fatal("one or more concurrent PING requests failed")
		}
	}
}
