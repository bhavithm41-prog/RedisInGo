package resp

import (
	"fmt"
	"strconv"
	"strings"
)

// SimpleString encodes s as a RESP Simple String, e.g. "OK" becomes
// "+OK\r\n". Simple strings must not contain \r or \n themselves —
// use BulkString for arbitrary content instead.
func SimpleString(s string) string {
	return "+" + s + "\r\n"
}

// Error encodes msg as a RESP Error reply, e.g. "-ERR bad thing\r\n".
func Error(msg string) string {
	return "-" + msg + "\r\n"
}

// Integer encodes n as a RESP Integer reply, e.g. ":42\r\n".
func Integer(n int) string {
	return ":" + strconv.Itoa(n) + "\r\n"
}

// BulkString encodes s as a RESP Bulk String, length-prefixed so it
// can safely contain any bytes, including \r or \n.
// Example: "hello" becomes "$5\r\nhello\r\n".
func BulkString(s string) string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(s), s)
}

// NullBulkString represents a missing value (e.g. GET on a key that
// doesn't exist) — distinct from an empty string. Real Redis clients
// decode this specifically as "nil", not "".
func NullBulkString() string {
	return "$-1\r\n"
}

// Array encodes a slice of already-encoded RESP elements as a RESP
// Array. Each element in elements must already be a fully-encoded
// RESP value (e.g. produced by BulkString or Integer) — Array just
// wraps them with the correct count prefix.
// Example: Array([]string{BulkString("a"), BulkString("b")})
// produces "*2\r\n$1\r\na\r\n$1\r\nb\r\n".
func Array(elements []string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "*%d\r\n", len(elements))
	for _, e := range elements {
		b.WriteString(e)
	}
	return b.String()
}

// NullArray represents a missing/empty array-type result in
// contexts where distinguishing "no result" from "empty result" is
// meaningful. Encoded as "*-1\r\n".
func NullArray() string {
	return "*-1\r\n"
}

// StringArray is a convenience helper: encodes a slice of plain Go
// strings as a RESP Array of Bulk Strings in one step.
// Example: StringArray([]string{"a", "b"}) produces the same output
// as Array([]string{BulkString("a"), BulkString("b")}).
func StringArray(items []string) string {
	elements := make([]string, len(items))
	for i, item := range items {
		elements[i] = BulkString(item)
	}
	return Array(elements)
}
