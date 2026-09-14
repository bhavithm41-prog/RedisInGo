package command

import (
	"strconv"
	"strings"

	"github.com/bhavithm41-prog/gocachedb/internal/resp"
	"github.com/bhavithm41-prog/gocachedb/internal/store"
)

type Handler struct {
	store    *store.Store
	dataFile string
}

func New(s *store.Store, dataFile string) *Handler {
	return &Handler{store: s, dataFile: dataFile}
}

// Execute parses a raw line of text into a command and arguments,
// runs the command against the store, and returns a fully
// RESP-encoded reply string (including its own trailing \r\n) —
// the caller should write this directly, with no extra newline.
func (h *Handler) Execute(line string) string {
	cmd, args := parseLine(line)
	if cmd == "" {
		return resp.Error("ERR empty command")
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
	case "SADD":
		return h.cmdSAdd(args)
	case "SREM":
		return h.cmdSRem(args)
	case "SISMEMBER":
		return h.cmdSIsMember(args)
	case "SMEMBERS":
		return h.cmdSMembers(args)
	case "HSET":
		return h.cmdHSet(args)
	case "HGET":
		return h.cmdHGet(args)
	case "HGETALL":
		return h.cmdHGetAll(args)
	case "HDEL":
		return h.cmdHDel(args)
	case "EXPIRE":
		return h.cmdExpire(args)
	case "TTL":
		return h.cmdTTL(args)
	case "SETEX":
		return h.cmdSetEx(args)
	case "SAVE":
		return h.cmdSave(args)
	case "BGSAVE":
		return h.cmdBgSave(args)
	default:
		return resp.Error("ERR unknown command '" + cmd + "'")
	}
}

// ---------- String / generic commands ----------

func (h *Handler) cmdPing(args []string) string {
	if len(args) != 0 {
		return resp.Error("ERR wrong number of arguments for 'PING'")
	}
	return resp.SimpleString("PONG")
}

func (h *Handler) cmdSet(args []string) string {
	if len(args) != 2 {
		return resp.Error("ERR wrong number of arguments for 'SET'")
	}
	h.store.Set(args[0], args[1])
	return resp.SimpleString("OK")
}

func (h *Handler) cmdGet(args []string) string {
	if len(args) != 1 {
		return resp.Error("ERR wrong number of arguments for 'GET'")
	}
	value, exists, err := h.store.Get(args[0])
	if err != nil {
		return resp.Error("ERR " + err.Error())
	}
	if !exists {
		return resp.NullBulkString()
	}
	return resp.BulkString(value)
}

func (h *Handler) cmdDel(args []string) string {
	if len(args) != 1 {
		return resp.Error("ERR wrong number of arguments for 'DEL'")
	}
	deleted := h.store.Del(args[0])
	if deleted {
		return resp.Integer(1)
	}
	return resp.Integer(0)
}

func (h *Handler) cmdExists(args []string) string {
	if len(args) != 1 {
		return resp.Error("ERR wrong number of arguments for 'EXISTS'")
	}
	if h.store.Exists(args[0]) {
		return resp.Integer(1)
	}
	return resp.Integer(0)
}

func (h *Handler) cmdKeys(args []string) string {
	if len(args) != 0 {
		return resp.Error("ERR wrong number of arguments for 'KEYS'")
	}
	return resp.StringArray(h.store.Keys())
}

func (h *Handler) cmdDBSize(args []string) string {
	if len(args) != 0 {
		return resp.Error("ERR wrong number of arguments for 'DBSIZE'")
	}
	return resp.Integer(len(h.store.Keys()))
}

func (h *Handler) cmdFlushDB(args []string) string {
	if len(args) != 0 {
		return resp.Error("ERR wrong number of arguments for 'FLUSHDB'")
	}
	for _, k := range h.store.Keys() {
		h.store.Del(k)
	}
	return resp.SimpleString("OK")
}

// ---------- List commands ----------

func (h *Handler) cmdLPush(args []string) string {
	if len(args) < 2 {
		return resp.Error("ERR wrong number of arguments for 'LPUSH'")
	}
	length, err := h.store.LPush(args[0], args[1:]...)
	if err != nil {
		return resp.Error("ERR " + err.Error())
	}
	return resp.Integer(length)
}

func (h *Handler) cmdRPush(args []string) string {
	if len(args) < 2 {
		return resp.Error("ERR wrong number of arguments for 'RPUSH'")
	}
	length, err := h.store.RPush(args[0], args[1:]...)
	if err != nil {
		return resp.Error("ERR " + err.Error())
	}
	return resp.Integer(length)
}

func (h *Handler) cmdLPop(args []string) string {
	if len(args) != 1 {
		return resp.Error("ERR wrong number of arguments for 'LPOP'")
	}
	value, popped, err := h.store.LPop(args[0])
	if err != nil {
		return resp.Error("ERR " + err.Error())
	}
	if !popped {
		return resp.NullBulkString()
	}
	return resp.BulkString(value)
}

func (h *Handler) cmdRPop(args []string) string {
	if len(args) != 1 {
		return resp.Error("ERR wrong number of arguments for 'RPOP'")
	}
	value, popped, err := h.store.RPop(args[0])
	if err != nil {
		return resp.Error("ERR " + err.Error())
	}
	if !popped {
		return resp.NullBulkString()
	}
	return resp.BulkString(value)
}

func (h *Handler) cmdLRange(args []string) string {
	if len(args) != 3 {
		return resp.Error("ERR wrong number of arguments for 'LRANGE'")
	}
	start, err1 := strconv.Atoi(args[1])
	stop, err2 := strconv.Atoi(args[2])
	if err1 != nil || err2 != nil {
		return resp.Error("ERR value is not an integer or out of range")
	}

	result, err := h.store.LRange(args[0], start, stop)
	if err != nil {
		return resp.Error("ERR " + err.Error())
	}
	return resp.StringArray(result)
}

// ---------- Set commands ----------

func (h *Handler) cmdSAdd(args []string) string {
	if len(args) < 2 {
		return resp.Error("ERR wrong number of arguments for 'SADD'")
	}
	added, err := h.store.SAdd(args[0], args[1:]...)
	if err != nil {
		return resp.Error("ERR " + err.Error())
	}
	return resp.Integer(added)
}

func (h *Handler) cmdSRem(args []string) string {
	if len(args) < 2 {
		return resp.Error("ERR wrong number of arguments for 'SREM'")
	}
	removed, err := h.store.SRem(args[0], args[1:]...)
	if err != nil {
		return resp.Error("ERR " + err.Error())
	}
	return resp.Integer(removed)
}

func (h *Handler) cmdSIsMember(args []string) string {
	if len(args) != 2 {
		return resp.Error("ERR wrong number of arguments for 'SISMEMBER'")
	}
	isMember, err := h.store.SIsMember(args[0], args[1])
	if err != nil {
		return resp.Error("ERR " + err.Error())
	}
	if isMember {
		return resp.Integer(1)
	}
	return resp.Integer(0)
}

func (h *Handler) cmdSMembers(args []string) string {
	if len(args) != 1 {
		return resp.Error("ERR wrong number of arguments for 'SMEMBERS'")
	}
	members, err := h.store.SMembers(args[0])
	if err != nil {
		return resp.Error("ERR " + err.Error())
	}
	return resp.StringArray(members)
}

// ---------- Hash commands ----------

func (h *Handler) cmdHSet(args []string) string {
	if len(args) != 3 {
		return resp.Error("ERR wrong number of arguments for 'HSET'")
	}
	created, err := h.store.HSet(args[0], args[1], args[2])
	if err != nil {
		return resp.Error("ERR " + err.Error())
	}
	if created {
		return resp.Integer(1)
	}
	return resp.Integer(0)
}

func (h *Handler) cmdHGet(args []string) string {
	if len(args) != 2 {
		return resp.Error("ERR wrong number of arguments for 'HGET'")
	}
	value, exists, err := h.store.HGet(args[0], args[1])
	if err != nil {
		return resp.Error("ERR " + err.Error())
	}
	if !exists {
		return resp.NullBulkString()
	}
	return resp.BulkString(value)
}

func (h *Handler) cmdHGetAll(args []string) string {
	if len(args) != 1 {
		return resp.Error("ERR wrong number of arguments for 'HGETALL'")
	}
	pairs, err := h.store.HGetAll(args[0])
	if err != nil {
		return resp.Error("ERR " + err.Error())
	}
	return resp.StringArray(pairs)
}

func (h *Handler) cmdHDel(args []string) string {
	if len(args) != 2 {
		return resp.Error("ERR wrong number of arguments for 'HDEL'")
	}
	deleted, err := h.store.HDel(args[0], args[1])
	if err != nil {
		return resp.Error("ERR " + err.Error())
	}
	if deleted {
		return resp.Integer(1)
	}
	return resp.Integer(0)
}

// ---------- Expiration commands ----------

func (h *Handler) cmdExpire(args []string) string {
	if len(args) != 2 {
		return resp.Error("ERR wrong number of arguments for 'EXPIRE'")
	}
	seconds, err := strconv.Atoi(args[1])
	if err != nil {
		return resp.Error("ERR value is not an integer or out of range")
	}
	set := h.store.Expire(args[0], seconds)
	if set {
		return resp.Integer(1)
	}
	return resp.Integer(0)
}

func (h *Handler) cmdTTL(args []string) string {
	if len(args) != 1 {
		return resp.Error("ERR wrong number of arguments for 'TTL'")
	}
	ttl := h.store.TTL(args[0])
	return resp.Integer(ttl)
}

func (h *Handler) cmdSetEx(args []string) string {
	if len(args) != 3 {
		return resp.Error("ERR wrong number of arguments for 'SETEX'")
	}
	seconds, err := strconv.Atoi(args[1])
	if err != nil {
		return resp.Error("ERR value is not an integer or out of range")
	}
	h.store.SetEx(args[0], seconds, args[2])
	return resp.SimpleString("OK")
}

// ---------- Persistence commands ----------

func (h *Handler) cmdSave(args []string) string {
	if len(args) != 0 {
		return resp.Error("ERR wrong number of arguments for 'SAVE'")
	}
	if err := h.store.SaveToFile(h.dataFile); err != nil {
		return resp.Error("ERR " + err.Error())
	}
	return resp.SimpleString("OK")
}

func (h *Handler) cmdBgSave(args []string) string {
	if len(args) != 0 {
		return resp.Error("ERR wrong number of arguments for 'BGSAVE'")
	}
	go func() {
		_ = h.store.SaveToFile(h.dataFile)
	}()
	return resp.SimpleString("Background saving started")
}
