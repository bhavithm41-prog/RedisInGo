package store

import (
	"errors"
	"sync"
)

// ErrWrongType is returned when a command is used against a key
// that holds a value of a different, incompatible type.
var ErrWrongType = errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")

// Store is a thread-safe in-memory key-value store.
// Values are stored as `any` so a single key can hold different
// underlying types: string, []string (list), or map[string]struct{} (set).
type Store struct {
	mu   sync.RWMutex
	data map[string]any
}

// New creates and returns a new, empty Store.
func New() *Store {
	return &Store{
		data: make(map[string]any),
	}
}

// ---------- String commands ----------

func (s *Store) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

func (s *Store) Get(key string) (string, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	raw, exists := s.data[key]
	if !exists {
		return "", false, nil
	}
	value, ok := raw.(string)
	if !ok {
		return "", true, ErrWrongType
	}
	return value, true, nil
}

// ---------- Generic key commands ----------

func (s *Store) Del(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, exists := s.data[key]
	if exists {
		delete(s.data, key)
	}
	return exists
}

func (s *Store) Exists(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.data[key]
	return exists
}

func (s *Store) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys
}

// ---------- List commands ----------

func (s *Store) LPush(key string, values ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	list, err := s.getOrCreateList(key)
	if err != nil {
		return 0, err
	}
	for _, v := range values {
		list = append([]string{v}, list...)
	}
	s.data[key] = list
	return len(list), nil
}

func (s *Store) RPush(key string, values ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	list, err := s.getOrCreateList(key)
	if err != nil {
		return 0, err
	}
	list = append(list, values...)
	s.data[key] = list
	return len(list), nil
}

func (s *Store) LPop(key string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	list, exists, err := s.getList(key)
	if err != nil {
		return "", false, err
	}
	if !exists || len(list) == 0 {
		return "", false, nil
	}
	value := list[0]
	list = list[1:]
	s.data[key] = list
	return value, true, nil
}

func (s *Store) RPop(key string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	list, exists, err := s.getList(key)
	if err != nil {
		return "", false, err
	}
	if !exists || len(list) == 0 {
		return "", false, nil
	}
	lastIdx := len(list) - 1
	value := list[lastIdx]
	list = list[:lastIdx]
	s.data[key] = list
	return value, true, nil
}

func (s *Store) LRange(key string, start, stop int) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list, exists, err := s.getList(key)
	if err != nil {
		return nil, err
	}
	if !exists || len(list) == 0 {
		return []string{}, nil
	}

	length := len(list)
	start = normalizeIndex(start, length)
	stop = normalizeIndex(stop, length)

	if start < 0 {
		start = 0
	}
	if stop >= length {
		stop = length - 1
	}
	if start > stop || start >= length {
		return []string{}, nil
	}

	result := make([]string, stop-start+1)
	copy(result, list[start:stop+1])
	return result, nil
}

// ---------- Set commands ----------

// SAdd adds one or more members to the set at key, creating the set
// if it doesn't exist. Returns the number of members that were
// newly added (members already present don't count).
func (s *Store) SAdd(key string, members ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	set, err := s.getOrCreateSet(key)
	if err != nil {
		return 0, err
	}

	added := 0
	for _, m := range members {
		if _, exists := set[m]; !exists {
			set[m] = struct{}{}
			added++
		}
	}
	s.data[key] = set
	return added, nil
}

// SRem removes one or more members from the set at key.
// Returns the number of members that were actually removed.
func (s *Store) SRem(key string, members ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	set, exists, err := s.getSet(key)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, nil
	}

	removed := 0
	for _, m := range members {
		if _, exists := set[m]; exists {
			delete(set, m)
			removed++
		}
	}
	s.data[key] = set
	return removed, nil
}

// SIsMember reports whether member is present in the set at key.
func (s *Store) SIsMember(key, member string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	set, exists, err := s.getSet(key)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}
	_, isMember := set[member]
	return isMember, nil
}

// SMembers returns all members of the set at key, in no
// guaranteed order (matching both Go's map iteration and real
// Redis's own unordered set semantics).
func (s *Store) SMembers(key string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	set, exists, err := s.getSet(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return []string{}, nil
	}

	members := make([]string, 0, len(set))
	for m := range set {
		members = append(members, m)
	}
	return members, nil
}

// ---------- Internal helpers ----------

func (s *Store) getList(key string) ([]string, bool, error) {
	raw, exists := s.data[key]
	if !exists {
		return nil, false, nil
	}
	list, ok := raw.([]string)
	if !ok {
		return nil, true, ErrWrongType
	}
	return list, true, nil
}

func (s *Store) getOrCreateList(key string) ([]string, error) {
	list, exists, err := s.getList(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return []string{}, nil
	}
	return list, nil
}

func (s *Store) getSet(key string) (map[string]struct{}, bool, error) {
	raw, exists := s.data[key]
	if !exists {
		return nil, false, nil
	}
	set, ok := raw.(map[string]struct{})
	if !ok {
		return nil, true, ErrWrongType
	}
	return set, true, nil
}

func (s *Store) getOrCreateSet(key string) (map[string]struct{}, error) {
	set, exists, err := s.getSet(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return make(map[string]struct{}), nil
	}
	return set, nil
}

func normalizeIndex(idx, length int) int {
	if idx < 0 {
		return length + idx
	}
	return idx
}
