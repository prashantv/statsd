package statsd

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

const maxPacketSize = 65536

// Metrics represents all the metrics collected by this statsd server.
type Metrics struct {
	sync.RWMutex

	counters map[string]int64
	gauges   map[string]int64
	timers   map[string][]time.Duration

	onUpdate []func()
}

func newMetrics() *Metrics {
	m := &Metrics{}
	m.initMaps()
	return m
}

func (m *Metrics) initMaps() {
	m.counters = make(map[string]int64)
	m.gauges = make(map[string]int64)
	m.timers = make(map[string][]time.Duration)
}

// Counters returns a copy of all the counters.
func (m *Metrics) Counters() map[string]int64 {
	m.RLock()
	defer m.RUnlock()

	return copyMapInt64(m.counters)
}

// Gauges returns a copy of all the gauges.
func (m *Metrics) Gauges() map[string]int64 {
	m.RLock()
	defer m.RUnlock()

	return copyMapInt64(m.gauges)
}

// Timers returns a copy of all the timers.
func (m *Metrics) Timers() map[string][]time.Duration {
	m.RLock()
	defer m.RUnlock()

	return copyMapDuration(m.timers)
}

// FlushAndSnapshot returns a snapshot of all the metrics and flushes the current metrics object.
// This is done atomically, and is optimized to avoid copying the full map.
func (m *Metrics) FlushAndSnapshot() *Snapshot {
	m.Lock()
	defer m.Unlock()

	// We can return the underlying maps since the flush means we no longer use this map.
	ss := &Snapshot{
		Counters: m.counters,
		Gauges:   m.gauges,
		Timers:   m.timers,
	}
	m.initMaps()
	return ss
}

// Snapshot returns a snapshot of all metrics without flushing them.
func (m *Metrics) Snapshot() *Snapshot {
	m.RLock()
	defer m.RUnlock()

	return &Snapshot{
		Counters: copyMapInt64(m.counters),
		Gauges:   copyMapInt64(m.gauges),
		Timers:   copyMapDuration(m.timers),
	}
}

// AddOnUpdate adds an event handler that is called on updates.
// It is called after processing a packet.
func (m *Metrics) AddOnUpdate(f func()) {
	m.Lock()
	m.onUpdate = append(m.onUpdate, f)
	m.Unlock()
}

func (m *Metrics) callOnUpdate() {
	m.RLock()
	updates := m.onUpdate
	m.RUnlock()
	for _, f := range updates {
		f()
	}
}

func (m *Metrics) processCounter(name, value []byte) error {
	valueInt, err := strconv.ParseInt(string(value), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid counter value %s: %v", value, err)
	}

	m.Lock()
	m.counters[string(name)] += valueInt
	m.Unlock()
	return nil
}

func (m *Metrics) processGauge(name, value []byte) error {
	valueInt, err := strconv.ParseInt(string(value), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid counter value %s: %v", value, err)
	}

	m.Lock()
	m.gauges[string(name)] = valueInt
	m.Unlock()
	return nil
}

func (m *Metrics) processTimer(name, value []byte) error {
	valueFloat, err := strconv.ParseFloat(string(value), 64)
	if err != nil {
		return fmt.Errorf("invalid counter value %s: %v", value, err)
	}

	sName := string(name)

	m.Lock()
	m.timers[sName] = append(m.timers[sName], time.Duration(valueFloat*float64(time.Millisecond)))
	m.Unlock()

	return nil
}
