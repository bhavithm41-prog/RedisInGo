package store

import (
	"fmt"
	"sync"
	"testing"

	"github.com/bhavithm41-prog/gocachedb/internal/metrics"
)

func TestConcurrentAccess(t *testing.T) {
	s := New(1000, metrics.New())

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

func TestLRUEviction(t *testing.T) {
	s := New(3, metrics.New())

	s.Set("a", "1")
	s.Set("b", "2")
	s.Set("c", "3")

	_, _, _ = s.Get("a")

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
