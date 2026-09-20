package store

import (
	"fmt"
	"testing"

	"github.com/bhavithm41-prog/gocachedb/internal/metrics"
)

// BenchmarkSet measures the cost of a single SET operation,
// run b.N times sequentially (single-threaded).
func BenchmarkSet(b *testing.B) {
	s := New(1000000, metrics.New()) // large capacity: benchmark SET itself, not eviction
	b.ResetTimer()                   // exclude setup time (store creation) from the measurement

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i%1000)
		s.Set(key, "some-benchmark-value")
	}
}

// BenchmarkGet measures the cost of a single GET operation on an
// existing key, run b.N times sequentially.
func BenchmarkGet(b *testing.B) {
	s := New(1000000, metrics.New())
	for i := 0; i < 1000; i++ {
		s.Set(fmt.Sprintf("key-%d", i), "some-benchmark-value")
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i%1000)
		_, _, _ = s.Get(key)
	}
}

// BenchmarkSetParallel measures SET throughput under real
// concurrent load from many goroutines simultaneously, revealing
// the actual cost of lock contention on the store's sync.RWMutex.
func BenchmarkSetParallel(b *testing.B) {
	s := New(1000000, metrics.New())
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key-%d", i%1000)
			s.Set(key, "some-benchmark-value")
			i++
		}
	})
}

// BenchmarkGetParallel measures GET throughput under concurrent
// load — since Get takes the full write lock (due to lazy
// expiration checks, per our Phase 6 design tradeoff), this
// reveals the real cost of that design decision under contention.
func BenchmarkGetParallel(b *testing.B) {
	s := New(1000000, metrics.New())
	for i := 0; i < 1000; i++ {
		s.Set(fmt.Sprintf("key-%d", i), "some-benchmark-value")
	}
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key-%d", i%1000)
			_, _, _ = s.Get(key)
			i++
		}
	})
}

// BenchmarkLPush measures list insertion cost. Recall from Phase 5:
// LPUSH is O(n) with our []string-slice-based implementation, since
// prepending requires shifting every existing element. This
// benchmark makes that real cost visible and measurable, rather
// than just theoretical.
func BenchmarkLPush(b *testing.B) {
	s := New(1000000, metrics.New())
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.LPush("mylist", "value")
	}
}

// BenchmarkRPush measures list append cost — expected to be
// meaningfully faster than BenchmarkLPush at scale, since RPUSH is
// O(1) amortized with a slice, unlike LPUSH's O(n) shift.
func BenchmarkRPush(b *testing.B) {
	s := New(1000000, metrics.New())
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.RPush("mylist", "value")
	}
}
