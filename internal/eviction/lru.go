package eviction

type node struct {
	key  string
	prev *node
	next *node
}

type LRU struct {
	capacity int
	items    map[string]*node
	head     *node
	tail     *node
}

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

func (l *LRU) Touch(key string) (evictedKey string, evicted bool) {
	if n, exists := l.items[key]; exists {
		l.moveToFront(n)
		return "", false
	}

	n := &node{key: key}
	l.items[key] = n
	l.insertAtFront(n)

	if len(l.items) > l.capacity {
		lru := l.tail.prev
		l.removeNode(lru)
		delete(l.items, lru.key)
		return lru.key, true
	}

	return "", false
}

func (l *LRU) Remove(key string) bool {
	n, exists := l.items[key]
	if !exists {
		return false
	}
	l.removeNode(n)
	delete(l.items, key)
	return true
}

func (l *LRU) Len() int {
	return len(l.items)
}

func (l *LRU) removeNode(n *node) {
	n.prev.next = n.next
	n.next.prev = n.prev
}

func (l *LRU) insertAtFront(n *node) {
	n.prev = l.head
	n.next = l.head.next
	l.head.next.prev = n
	l.head.next = n
}

func (l *LRU) moveToFront(n *node) {
	l.removeNode(n)
	l.insertAtFront(n)
}
