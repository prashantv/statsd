package main

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/prashantv/statsd"
)

// Snapshot takes metrics in a given window and records a snapshot.
type Snapshot struct {
	Index    int                      `json:"index"`
	Time     string                   `json:"key"`
	Counters map[string]int64         `json:"counters"`
	Gauges   map[string]int64         `json:"gauges"`
	Timers   map[string]TimerSnapshot `json:"timers"`
}

// TimerSnapshot takes timers in a given window and creates a snapshot.
type TimerSnapshot struct {
	// TODO: store a list of quantiles, and allow merging.
	P50 durationMS `json:"p50"`
	P90 durationMS `json:"p90"`
	P95 durationMS `json:"p95"`
	P99 durationMS `json:"p99"`
}

var (
	snapshotsLock sync.RWMutex
	snapshots     []Snapshot
)

func fakeData(t time.Time) Snapshot {
	// Fake counters
	s := Snapshot{
		Counters: make(map[string]int64),
		Timers:   make(map[string]TimerSnapshot),
		Time:     t.Format("15:04:05"),
	}
	for i := 0; i < 2; i++ {
		s.Counters[fmt.Sprintf("counter-%v", i)] = rand.Int63n(100)
	}

	max := int64(time.Second)
	s.Timers["tmr"] = TimerSnapshot{
		P50: durationMS(rand.Int63n(max) / 2),
		P90: durationMS(rand.Int63n(max) * 2 / 3),
		P95: durationMS(rand.Int63n(max) * 3 / 4),
		P99: durationMS(rand.Int63n(max) * 5 / 6),
	}

	return s
}

func recordMetrics(t time.Time, m *statsd.Metrics) {
	snapshot := toSnapshot(t, m.FlushAndSnapshot())
	snapshot = fakeData(t)

	snapshotsLock.Lock()
	snapshot.Index = len(snapshots)
	snapshots = append(snapshots, snapshot)
	snapshotsLock.Unlock()
}

func toSnapshot(t time.Time, m *statsd.Snapshot) Snapshot {
	timers := make(map[string]TimerSnapshot, len(m.Timers))
	for k, v := range m.Timers {
		timers[k] = toTimerSnapshot(v)
	}

	return Snapshot{
		Time:     t.Format("15:04:05"),
		Counters: m.Counters,
		Gauges:   m.Gauges,
		Timers:   timers,
	}
}
func toTimerSnapshot(timers []time.Duration) TimerSnapshot {
	sort.Sort(byDuration(timers))
	return TimerSnapshot{
		P50: durationMS(getQuantile(timers, 0.5)),
		P90: durationMS(getQuantile(timers, 0.9)),
		P95: durationMS(getQuantile(timers, 0.95)),
		P99: durationMS(getQuantile(timers, 0.99)),
	}
}
