package eviction

// node is one entry in the doubly linked list. It stores both the
// key and value so that, when we evict the tail node, we know which
// key to also remove from the hash map — not just which value to drop.
type node struct {
	key   string
	value string
	prev  *node
	next  *node
}

// LRU is a fixed-capacity Least-Recently-Used cache.
// It combines a hash map (for O(1) key lookup) with a doubly linked
// list (for O(1) reordering and eviction) to achieve O(1) Get and
// Put operations.
//
// NOTE: this type is NOT thread-safe on its own — concurrency
// protection is added when we integrate it into Store in the next
// sub-step, consistent with how Store already handles locking.
type LRU struct {
	capacity int
	items    map[string]*node

	// head and tail are sentinel (dummy) nodes that never hold real
	// data. The real list lives strictly between them:
	//   head <-> mostRecent <-> ... <-> leastRecent <-> tail
	// This avoids nil-pointer special cases for an empty list or a
	// list with only one real node.
	head *node
	tail *node
}

// New creates a new LRU cache with the given maximum capacity.
// Capacity must be at least 1.
func New(capacity int) *LRU {
	head := &node{}
	tail := &node{}
	head.next = tail
	tail.prev = head

	return &LRU{
		capacity: capacity,
		items:    make(map[string]*node),
		head:     head,
		tail:     tail,
	}
}

// Get retrieves the value for key, if present, and marks it as the
// most recently used item. The second return value reports whether
// the key was found.
func (l *LRU) Get(key string) (string, bool) {
	n, exists := l.items[key]
	if !exists {
		return "", false
	}
	l.moveToFront(n)
	return n.value, true
}

// Put inserts or updates the value for key, marking it as the most
// recently used item. If this causes the cache to exceed its
// capacity, the least recently used item is evicted. Put returns
// the key that was evicted, and true, if an eviction occurred.
func (l *LRU) Put(key, value string) (evictedKey string, evicted bool) {
	if n, exists := l.items[key]; exists {
		n.value = value
		l.moveToFront(n)
		return "", false
	}

	n := &node{key: key, value: value}
	l.items[key] = n
	l.insertAtFront(n)

	if len(l.items) > l.capacity {
		lru := l.tail.prev // the real node just before the tail sentinel
		l.removeNode(lru)
		delete(l.items, lru.key)
		return lru.key, true
	}

	return "", false
}

// Remove deletes key from the cache entirely, if present.
// Returns true if the key existed and was removed.
func (l *LRU) Remove(key string) bool {
	n, exists := l.items[key]
	if !exists {
		return false
	}
	l.removeNode(n)
	delete(l.items, key)
	return true
}

// Len returns the current number of items in the cache.
func (l *LRU) Len() int {
	return len(l.items)
}

// ---------- Internal doubly-linked-list helpers ----------

// removeNode splices n out of the list by re-linking its neighbors
// to each other. n itself is left with dangling prev/next pointers,
// which is fine since the caller is about to discard it or reinsert
// it elsewhere via insertAtFront.
func (l *LRU) removeNode(n *node) {
	n.prev.next = n.next
	n.next.prev = n.prev
}

// insertAtFront splices n in immediately after the head sentinel,
// making it the most recently used node.
func (l *LRU) insertAtFront(n *node) {
	n.prev = l.head
	n.next = l.head.next
	l.head.next.prev = n
	l.head.next = n
}

// moveToFront removes n from its current position and reinserts it
// at the front — used whenever a node is accessed via Get, or
// updated via Put, to mark it as most recently used.
func (l *LRU) moveToFront(n *node) {
	l.removeNode(n)
	l.insertAtFront(n)
}
