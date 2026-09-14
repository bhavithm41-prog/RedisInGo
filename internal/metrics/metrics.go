package metrics

import (
	"fmt"
	"sync/atomic"
	"time"
)

// Metrics tracks operational counters for the server, safe for
// concurrent access from many goroutines via sync/atomic —
// a lighter-weight alternative to sync.RWMutex, well suited to
// simple independent counters like these.
type Metrics struct {
	startTime        time.Time
	totalCommands    atomic.Int64
	getCount         atomic.Int64
	setCount         atomic.Int64
	cacheHits        atomic.Int64
	cacheMisses      atomic.Int64
	connectedClients atomic.Int64
}

// New creates a Metrics tracker with its start time set to now.
func New() *Metrics {
	return &Metrics{startTime: time.Now()}
}

func (m *Metrics) IncrCommands() { m.totalCommands.Add(1) }
func (m *Metrics) IncrGet()      { m.getCount.Add(1) }
func (m *Metrics) IncrSet()      { m.setCount.Add(1) }
func (m *Metrics) IncrHit()      { m.cacheHits.Add(1) }
func (m *Metrics) IncrMiss()     { m.cacheMisses.Add(1) }
func (m *Metrics) ClientConnected()    { m.connectedClients.Add(1) }
func (m *Metrics) ClientDisconnected() { m.connectedClients.Add(-1) }

// Snapshot returns a human-readable report of current metrics,
// formatted similarly to real Redis's own INFO command output.
func (m *Metrics) Snapshot(evictions int, dbSize int) string {
	uptime := time.Since(m.startTime).Round(time.Second)

	hits := m.cacheHits.Load()
	misses := m.cacheMisses.Load()
	hitRatio := 0.0
	if hits+misses > 0 {
		hitRatio = float64(hits) / float64(hits+misses) * 100
	}

	return fmt.Sprintf(
		"uptime_seconds:%d\r\n"+
			"connected_clients:%d\r\n"+
			"total_commands_processed:%d\r\n"+
			"total_get_commands:%d\r\n"+
			"total_set_commands:%d\r\n"+
			"keyspace_hits:%d\r\n"+
			"keyspace_misses:%d\r\n"+
			"keyspace_hit_ratio_percent:%.2f\r\n"+
			"evicted_keys:%d\r\n"+
			"db_size:%d\r\n",
		int64(uptime.Seconds()),
		m.connectedClients.Load(),
		m.totalCommands.Load(),
		m.getCount.Load(),
		m.setCount.Load(),
		hits,
		misses,
		hitRatio,
		evictions,
		dbSize,
	)
}
