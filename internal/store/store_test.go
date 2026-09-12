package store

import (
	"fmt"
	"sync"
	"testing"
)

// TestConcurrentAccess hammers the store with many goroutines doing
// simultaneous reads and writes. Run with -race to detect data races:
//
//   go test -race ./internal/store/...
func TestConcurrentAccess(t *testing.T) {
	s := New()

	var wg sync.WaitGroup

	// Spin up 50 goroutines, each doing 100 SET/GET operations
	// on the SAME shared store, all at the same time.
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				key := fmt.Sprintf("key-%d", id)
				s.Set(key, "value")
				_, _ = s.Get(key)
				_ = s.Exists(key)
			}
		}(i)
	}

	wg.Wait()
}
