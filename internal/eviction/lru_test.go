package eviction

import "testing"

func TestBasicGetPut(t *testing.T) {
	l := New(2)

	l.Put("a", "1")
	l.Put("b", "2")

	if v, ok := l.Get("a"); !ok || v != "1" {
		t.Fatalf("expected a=1, got v=%q ok=%v", v, ok)
	}
	if v, ok := l.Get("b"); !ok || v != "2" {
		t.Fatalf("expected b=2, got v=%q ok=%v", v, ok)
	}
}

func TestEvictionOrder(t *testing.T) {
	l := New(2)

	l.Put("a", "1")
	l.Put("b", "2")
	// Cache is now full: [b, a] from most- to least-recently-used.

	// Access "a" so it becomes most recently used, making "b" the LRU item.
	l.Get("a")

	// Inserting "c" should evict "b" (the least recently used), not "a".
	evictedKey, evicted := l.Put("c", "3")
	if !evicted || evictedKey != "b" {
		t.Fatalf("expected eviction of 'b', got evictedKey=%q evicted=%v", evictedKey, evicted)
	}

	if _, ok := l.Get("b"); ok {
		t.Fatalf("expected 'b' to have been evicted, but it's still present")
	}
	if v, ok := l.Get("a"); !ok || v != "1" {
		t.Fatalf("expected 'a' to still be present with value 1, got v=%q ok=%v", v, ok)
	}
	if v, ok := l.Get("c"); !ok || v != "3" {
		t.Fatalf("expected 'c' to be present with value 3, got v=%q ok=%v", v, ok)
	}
}

func TestUpdateExistingKeyDoesNotEvict(t *testing.T) {
	l := New(2)

	l.Put("a", "1")
	l.Put("b", "2")

	// Updating an existing key should NOT count as a new insertion,
	// and should NOT trigger eviction.
	evictedKey, evicted := l.Put("a", "100")
	if evicted {
		t.Fatalf("did not expect eviction when updating existing key, got evictedKey=%q", evictedKey)
	}

	if v, ok := l.Get("a"); !ok || v != "100" {
		t.Fatalf("expected a=100 after update, got v=%q ok=%v", v, ok)
	}
	if l.Len() != 2 {
		t.Fatalf("expected length 2, got %d", l.Len())
	}
}

func TestRemove(t *testing.T) {
	l := New(2)

	l.Put("a", "1")
	l.Put("b", "2")

	if removed := l.Remove("a"); !removed {
		t.Fatalf("expected Remove(a) to return true")
	}
	if _, ok := l.Get("a"); ok {
		t.Fatalf("expected 'a' to be gone after Remove")
	}
	if l.Len() != 1 {
		t.Fatalf("expected length 1 after removing one item, got %d", l.Len())
	}

	if removed := l.Remove("nonexistent"); removed {
		t.Fatalf("expected Remove of missing key to return false")
	}
}
