package store

import "sync"

// Store is a thread-safe in-memory key-value store.
// A sync.RWMutex protects the underlying map: multiple goroutines
// may read concurrently, but writes are exclusive.
type Store struct {
	mu   sync.RWMutex
	data map[string]string
}

// New creates and returns a new, empty Store.
func New() *Store {
	return &Store{
		data: make(map[string]string),
	}
}

// Set stores the given value under the given key.
// If the key already exists, its value is overwritten.
func (s *Store) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

// Get retrieves the value for a key.
// The second return value reports whether the key existed.
func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, exists := s.data[key]
	return value, exists
}

// Del removes a key from the store.
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

// Exists reports whether a key is present in the store.
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
