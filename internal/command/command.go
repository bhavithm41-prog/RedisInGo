package command

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bhavithm41-prog/gocachedb/internal/store"
)

type Handler struct {
	store *store.Store
}

func New(s *store.Store) *Handler {
	return &Handler{store: s}
}

func (h *Handler) Execute(line string) string {
	cmd, args := parseLine(line)
	if cmd == "" {
		return "ERR empty command"
	}
	return h.dispatch(cmd, args)
}

func parseLine(line string) (string, []string) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return "", nil
	}
	cmd := strings.ToUpper(fields[0])
	args := fields[1:]
	return cmd, args
}

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
	case "LPUSH":
		return h.cmdLPush(args)
	case "RPUSH":
		return h.cmdRPush(args)
	case "LPOP":
		return h.cmdLPop(args)
	case "RPOP":
		return h.cmdRPop(args)
	case "LRANGE":
		return h.cmdLRange(args)
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
	value, exists, err := h.store.Get(args[0])
	if err != nil {
		return "ERR " + err.Error()
	}
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

func (h *Handler) cmdLPush(args []string) string {
	if len(args) < 2 {
		return "ERR wrong number of arguments for 'LPUSH'"
	}
	length, err := h.store.LPush(args[0], args[1:]...)
	if err != nil {
		return "ERR " + err.Error()
	}
	return strconv.Itoa(length)
}

func (h *Handler) cmdRPush(args []string) string {
	if len(args) < 2 {
		return "ERR wrong number of arguments for 'RPUSH'"
	}
	length, err := h.store.RPush(args[0], args[1:]...)
	if err != nil {
		return "ERR " + err.Error()
	}
	return strconv.Itoa(length)
}

func (h *Handler) cmdLPop(args []string) string {
	if len(args) != 1 {
		return "ERR wrong number of arguments for 'LPOP'"
	}
	value, popped, err := h.store.LPop(args[0])
	if err != nil {
		return "ERR " + err.Error()
	}
	if !popped {
		return "(nil)"
	}
	return value
}

func (h *Handler) cmdRPop(args []string) string {
	if len(args) != 1 {
		return "ERR wrong number of arguments for 'RPOP'"
	}
	value, popped, err := h.store.RPop(args[0])
	if err != nil {
		return "ERR " + err.Error()
	}
	if !popped {
		return "(nil)"
	}
	return value
}

func (h *Handler) cmdLRange(args []string) string {
	if len(args) != 3 {
		return "ERR wrong number of arguments for 'LRANGE'"
	}
	start, err1 := strconv.Atoi(args[1])
	stop, err2 := strconv.Atoi(args[2])
	if err1 != nil || err2 != nil {
		return "ERR value is not an integer or out of range"
	}

	result, err := h.store.LRange(args[0], start, stop)
	if err != nil {
		return "ERR " + err.Error()
	}
	if len(result) == 0 {
		return "(empty list)"
	}
	return strings.Join(result, " ")
}
