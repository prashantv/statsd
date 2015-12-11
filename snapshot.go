package statsd

import "time"

// Snapshot of the metrics at a certain point in time.
type Snapshot struct {
	Counters map[string]int64
	Gauges   map[string]int64
	Timers   map[string][]time.Duration
}
