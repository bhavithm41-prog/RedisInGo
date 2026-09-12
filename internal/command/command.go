package command

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bhavithm41-prog/gocachedb/internal/store"
)

// Handler holds a reference to the store so it can execute commands
// against it. This keeps command logic separate from the raw
// networking code in the server package.
type Handler struct {
	store *store.Store
}

// New creates a new command Handler backed by the given store.
func New(s *store.Store) *Handler {
	return &Handler{store: s}
}

// Execute parses a raw line of text into a command and arguments,
// runs the command against the store, and returns a response string
// (never including the trailing newline — the caller adds that).
func (h *Handler) Execute(line string) string {
	cmd, args := parseLine(line)
	if cmd == "" {
		return "ERR empty command"
	}
	return h.dispatch(cmd, args)
}

// parseLine splits a raw line into an uppercased command name and
// its arguments. strings.Fields handles multiple/irregular spaces
// gracefully, unlike a plain strings.Split.
func parseLine(line string) (string, []string) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return "", nil
	}
	cmd := strings.ToUpper(fields[0])
	args := fields[1:]
	return cmd, args
}

// dispatch executes a single parsed command against the store.
func (h *Handler) dispatch(cmd string, args []string) string {
	switch cmd {
	case "PING":
		return h.cmdPing(args)
	case "SET":
		return h.cmdSet(args)
	case "GET":
		return h.cmdGet(args)
	case "DEL":
		return h.cmdDel(args)
	case "EXISTS":
		return h.cmdExists(args)
	case "KEYS":
		return h.cmdKeys(args)
	case "DBSIZE":
		return h.cmdDBSize(args)
	case "FLUSHDB":
		return h.cmdFlushDB(args)
	default:
		return fmt.Sprintf("ERR unknown command '%s'", cmd)
	}
}

func (h *Handler) cmdPing(args []string) string {
	if len(args) != 0 {
		return "ERR wrong number of arguments for 'PING'"
	}
	return "PONG"
}

func (h *Handler) cmdSet(args []string) string {
	if len(args) != 2 {
		return "ERR wrong number of arguments for 'SET'"
	}
	h.store.Set(args[0], args[1])
	return "OK"
}

func (h *Handler) cmdGet(args []string) string {
	if len(args) != 1 {
		return "ERR wrong number of arguments for 'GET'"
	}
	value, exists := h.store.Get(args[0])
	if !exists {
		return "(nil)"
	}
	return value
}

func (h *Handler) cmdDel(args []string) string {
	if len(args) != 1 {
		return "ERR wrong number of arguments for 'DEL'"
	}
	deleted := h.store.Del(args[0])
	if deleted {
		return "1"
	}
	return "0"
}

func (h *Handler) cmdExists(args []string) string {
	if len(args) != 1 {
		return "ERR wrong number of arguments for 'EXISTS'"
	}
	if h.store.Exists(args[0]) {
		return "1"
	}
	return "0"
}

func (h *Handler) cmdKeys(args []string) string {
	if len(args) != 0 {
		return "ERR wrong number of arguments for 'KEYS'"
	}
	keys := h.store.Keys()
	if len(keys) == 0 {
		return "(empty list)"
	}
	return strings.Join(keys, " ")
}

func (h *Handler) cmdDBSize(args []string) string {
	if len(args) != 0 {
		return "ERR wrong number of arguments for 'DBSIZE'"
	}
	return strconv.Itoa(len(h.store.Keys()))
}

func (h *Handler) cmdFlushDB(args []string) string {
	if len(args) != 0 {
		return "ERR wrong number of arguments for 'FLUSHDB'"
	}
	for _, k := range h.store.Keys() {
		h.store.Del(k)
	}
	return "OK"
}
