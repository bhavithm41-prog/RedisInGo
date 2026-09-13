package eviction

import "testing"

func TestBasicTouch(t *testing.T) {
	l := New(2)
	l.Touch("a")
	l.Touch("b")
	if l.Len() != 2 {
		t.Fatalf("expected length 2, got %d", l.Len())
	}
}

func TestEvictionOrder(t *testing.T) {
	l := New(2)
	l.Touch("a")
	l.Touch("b")
	l.Touch("a")

	evictedKey, evicted := l.Touch("c")
	if !evicted || evictedKey != "b" {
		t.Fatalf("expected eviction of 'b', got evictedKey=%q evicted=%v", evictedKey, evicted)
	}
	if l.Len() != 2 {
		t.Fatalf("expected length 2 after eviction, got %d", l.Len())
	}
}

func TestTouchExistingKeyDoesNotEvict(t *testing.T) {
	l := New(2)
	l.Touch("a")
	l.Touch("b")

	evictedKey, evicted := l.Touch("a")
	if evicted {
		t.Fatalf("did not expect eviction when re-touching existing key, got evictedKey=%q", evictedKey)
	}
	if l.Len() != 2 {
		t.Fatalf("expected length 2, got %d", l.Len())
	}
}

func TestRemove(t *testing.T) {
	l := New(2)
	l.Touch("a")
	l.Touch("b")

	if removed := l.Remove("a"); !removed {
		t.Fatalf("expected Remove(a) to return true")
	}
	if l.Len() != 1 {
		t.Fatalf("expected length 1 after removing one item, got %d", l.Len())
	}
	if removed := l.Remove("nonexistent"); removed {
		t.Fatalf("expected Remove of missing key to return false")
	}
}
