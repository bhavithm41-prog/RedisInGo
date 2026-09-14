package store

import (
	"fmt"
	"os"
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

// TestSaveAndLoadRoundTrip verifies that saving a store's full
// contents to disk and loading them into a fresh store produces
// an exact match — proving SaveToFile and LoadFromFile are correct
// inverses of each other across every data type.
func TestSaveAndLoadRoundTrip(t *testing.T) {
	original := New(100, metrics.New())

	original.Set("name", "Bhavith")
	original.RPush("fruits", "apple", "banana")
	original.SAdd("tags", "go", "backend")
	original.HSet("user:1", "role", "engineer")
	original.Expire("name", 100)

	path := t.TempDir() + "/snapshot.rdb"
	if err := original.SaveToFile(path); err != nil {
		t.Fatalf("SaveToFile failed: %v", err)
	}

	loaded := New(100, metrics.New())
	if err := loaded.LoadFromFile(path); err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	if v, exists, _ := loaded.Get("name"); !exists || v != "Bhavith" {
		t.Fatalf("loaded name = %q exists=%v, want Bhavith/true", v, exists)
	}

	fruits, err := loaded.LRange("fruits", 0, -1)
	if err != nil || len(fruits) != 2 || fruits[0] != "apple" || fruits[1] != "banana" {
		t.Fatalf("loaded fruits = %v err=%v, want [apple banana]", fruits, err)
	}

	if isMember, _ := loaded.SIsMember("tags", "go"); !isMember {
		t.Fatalf("expected 'go' to be a member of loaded tags set")
	}

	role, exists, _ := loaded.HGet("user:1", "role")
	if !exists || role != "engineer" {
		t.Fatalf("loaded user:1 role = %q exists=%v, want engineer/true", role, exists)
	}

	ttl := loaded.TTL("name")
	if ttl <= 0 || ttl > 100 {
		t.Fatalf("loaded TTL for name = %d, want a positive value <= 100", ttl)
	}
}

// TestLoadFromMissingFile verifies that loading a file that doesn't
// exist is treated as "start empty", not an error.
func TestLoadFromMissingFile(t *testing.T) {
	s := New(100, metrics.New())
	path := t.TempDir() + "/does-not-exist.rdb"

	if err := s.LoadFromFile(path); err != nil {
		t.Fatalf("LoadFromFile on missing file should not error, got: %v", err)
	}
	if len(s.Keys()) != 0 {
		t.Fatalf("expected empty store after loading missing file, got %d keys", len(s.Keys()))
	}
}

// TestLoadFromCorruptedFile verifies that loading a corrupted file
// returns an error without crashing, and does not wipe out any
// data the store already had in memory.
func TestLoadFromCorruptedFile(t *testing.T) {
	path := t.TempDir() + "/corrupted.rdb"
	if err := os.WriteFile(path, []byte("this is not a valid gob snapshot"), 0644); err != nil {
		t.Fatalf("failed to write corrupted test file: %v", err)
	}

	s := New(100, metrics.New())
	s.Set("existing", "data") // pre-existing in-memory data, before the failed load

	err := s.LoadFromFile(path)
	if err == nil {
		t.Fatal("expected an error loading a corrupted file, got nil")
	}

	// Existing in-memory data must survive a failed load attempt.
	if v, exists, _ := s.Get("existing"); !exists || v != "data" {
		t.Fatalf("existing data was lost after failed load: v=%q exists=%v", v, exists)
	}
}
