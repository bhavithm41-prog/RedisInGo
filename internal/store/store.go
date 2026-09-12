package store

import (
	"errors"
	"sync"
)

// ErrWrongType is returned when a command is used against a key
// that holds a value of a different, incompatible type — e.g.
// calling LPUSH on a key that currently holds a plain string.
var ErrWrongType = errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")

// Store is a thread-safe in-memory key-value store.
// Values are stored as `any` so a single key can hold different
// underlying types (string today; lists, sets, hashes in later phases).
// A sync.RWMutex protects the map: multiple goroutines may read
// concurrently, but writes are exclusive.
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

// Set stores a string value under the given key, overwriting
// whatever was there before (regardless of its previous type).
func (s *Store) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

// Get retrieves a string value for a key.
// Returns ErrWrongType if the key exists but holds a non-string value.
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

// ---------- Generic key commands (work on any type) ----------

// Del removes a key from the store, regardless of its value's type.
// It returns true if the key existed and was deleted.
func (s *Store) Del(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, exists := s.data[key]
	if exists {
		delete(s.data, key)
	}
	return exists
}

// Exists reports whether a key is present in the store, regardless
// of its value's type.
func (s *Store) Exists(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.data[key]
	return exists
}

// Keys returns a slice of all keys currently in the store.
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

// LPush inserts one or more values at the head (left) of the list
// stored at key, creating the list if the key doesn't exist yet.
// Values are pushed one at a time, so multiple values end up in
// reverse order at the head — this matches real Redis semantics.
// Returns the new length of the list, or ErrWrongType if the key
// holds a non-list value.
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

// RPush inserts one or more values at the tail (right) of the list
// stored at key, creating the list if the key doesn't exist yet.
// Returns the new length of the list, or ErrWrongType if the key
// holds a non-list value.
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

// LPop removes and returns the leftmost (first) element of the list
// at key. The second return value reports whether an element was
// popped (false if the key doesn't exist or the list is empty).
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

// RPop removes and returns the rightmost (last) element of the list
// at key. The second return value reports whether an element was
// popped (false if the key doesn't exist or the list is empty).
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

// LRange returns the elements of the list at key between start and
// stop, both inclusive. Negative indices count from the end of the
// list (-1 is the last element). Out-of-range indices are clamped,
// matching real Redis behavior, rather than erroring.
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

	// Copy the slice so callers can't mutate our internal list.
	result := make([]string, stop-start+1)
	copy(result, list[start:stop+1])
	return result, nil
}

// ---------- Internal helpers ----------

// getList fetches the list stored at key, if any.
// Returns (nil, false, nil) if the key doesn't exist,
// and (nil, true, ErrWrongType) if the key holds a non-list value.
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

// getOrCreateList fetches the list stored at key, creating an empty
// one if the key doesn't exist yet. Must be called with the write
// lock already held.
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

// normalizeIndex converts a possibly-negative Redis-style index
// into a real slice index. -1 means the last element, -2 the
// second-to-last, and so on.
func normalizeIndex(idx, length int) int {
	if idx < 0 {
		return length + idx
	}
	return idx
}
