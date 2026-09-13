package store

import (
	"fmt"
	"sync"
	"testing"
)

func TestConcurrentAccess(t *testing.T) {
	s := New(1000)

	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				key := fmt.Sprintf("key-%d", id)
				s.Set(key, "value")
				_, _, _ = s.Get(key)
				_ = s.Exists(key)
			}
		}(i)
	}

	wg.Wait()
}

// TestLRUEviction verifies that once the store exceeds its
// configured key capacity, the least recently used key is evicted
// — and that its actual data is really gone, not just untracked.
func TestLRUEviction(t *testing.T) {
	s := New(3) // capacity: only 3 keys allowed at once

	s.Set("a", "1")
	s.Set("b", "2")
	s.Set("c", "3")
	// Store is now full: [c, b, a] from most- to least-recently-used.

	// Access "a" so it becomes most recently used, making "b" the LRU key.
	_, _, _ = s.Get("a")

	// Inserting "d" should evict "b" (the least recently used), not "a" or "c".
	s.Set("d", "4")

	if _, exists, _ := s.Get("b"); exists {
		t.Fatalf("expected key 'b' to have been evicted, but it still exists")
	}

	if v, exists, _ := s.Get("a"); !exists || v != "1" {
		t.Fatalf("expected 'a' to still exist with value 1, got v=%q exists=%v", v, exists)
	}
	if v, exists, _ := s.Get("c"); !exists || v != "3" {
		t.Fatalf("expected 'c' to still exist with value 3, got v=%q exists=%v", v, exists)
	}
	if v, exists, _ := s.Get("d"); !exists || v != "4" {
		t.Fatalf("expected 'd' to exist with value 4, got v=%q exists=%v", v, exists)
	}

	if evictions := s.Evictions(); evictions != 1 {
		t.Fatalf("expected 1 total eviction, got %d", evictions)
	}
}
