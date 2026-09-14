package resp

import "testing"

func TestSimpleString(t *testing.T) {
	got := SimpleString("OK")
	want := "+OK\r\n"
	if got != want {
		t.Fatalf("SimpleString(OK) = %q, want %q", got, want)
	}
}

func TestError(t *testing.T) {
	got := Error("ERR bad thing")
	want := "-ERR bad thing\r\n"
	if got != want {
		t.Fatalf("Error(...) = %q, want %q", got, want)
	}
}

func TestInteger(t *testing.T) {
	got := Integer(42)
	want := ":42\r\n"
	if got != want {
		t.Fatalf("Integer(42) = %q, want %q", got, want)
	}

	gotNeg := Integer(-1)
	wantNeg := ":-1\r\n"
	if gotNeg != wantNeg {
		t.Fatalf("Integer(-1) = %q, want %q", gotNeg, wantNeg)
	}
}

func TestBulkString(t *testing.T) {
	got := BulkString("hello")
	want := "$5\r\nhello\r\n"
	if got != want {
		t.Fatalf("BulkString(hello) = %q, want %q", got, want)
	}

	gotEmpty := BulkString("")
	wantEmpty := "$0\r\n\r\n"
	if gotEmpty != wantEmpty {
		t.Fatalf("BulkString(\"\") = %q, want %q", gotEmpty, wantEmpty)
	}
}

func TestNullBulkString(t *testing.T) {
	got := NullBulkString()
	want := "$-1\r\n"
	if got != want {
		t.Fatalf("NullBulkString() = %q, want %q", got, want)
	}
}

func TestStringArray(t *testing.T) {
	got := StringArray([]string{"foo", "bar"})
	want := "*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"
	if got != want {
		t.Fatalf("StringArray([foo bar]) = %q, want %q", got, want)
	}
}

func TestStringArrayEmpty(t *testing.T) {
	got := StringArray([]string{})
	want := "*0\r\n"
	if got != want {
		t.Fatalf("StringArray([]) = %q, want %q", got, want)
	}
}

func TestNullArray(t *testing.T) {
	got := NullArray()
	want := "*-1\r\n"
	if got != want {
		t.Fatalf("NullArray() = %q, want %q", got, want)
	}
}
