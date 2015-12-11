package statsd

import (
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProtocolUDP(t *testing.T) {
	localAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	require.NoError(t, err, "ResolveUDPAddr failed")

	metrics, udpAddr, err := Start("127.0.0.1:0")
	require.NoError(t, err, "Failed to start statsd server")

	conn, err := net.DialUDP("udp", localAddr, udpAddr)
	require.NoError(t, err, "Failed to connect to statsd server")

	packets := []string{
		"c1:1|c",
		"c1:2|c\nt1:5.2|ms\ng1:3|g\ng2:4|g",
		"g1:5|g",
		"t1:4.8|ms\nc1:1|c",
	}
	expected := &Snapshot{
		Counters: map[string]int64{
			"c1": 4,
		},
		Gauges: map[string]int64{
			"g1": 5,
			"g2": 4,
		},
		Timers: map[string][]time.Duration{
			"t1": []time.Duration{5200 * time.Microsecond, 4800 * time.Microsecond},
		},
	}

	var wg sync.WaitGroup
	metrics.AddOnUpdate(func() {
		wg.Done()
	})
	wg.Add(len(packets))
	for _, packet := range packets {
		_, err := conn.Write([]byte(packet))
		require.NoError(t, err, "Failed to write packet to UDP connection")
	}

	// Wait till all the metrics are processed
	wg.Wait()

	got := metrics.FlushAndSnapshot()
	assert.Equal(t, expected, got, "Snapshot mismatch")
}

func TestProcessCounter(t *testing.T) {
	tests := []struct {
		values   []string
		expected int64
		wantErr  bool
	}{
		{
			values:   []string{"1"},
			expected: 1,
		},
		{
			values:   []string{"1", "2", "3"},
			expected: 6,
		},
		{
			values:  []string{"1.0"},
			wantErr: true,
		},
		{
			values:  []string{"abc"},
			wantErr: true,
		},
		{
			values:   []string{"1", "abc", "9"},
			wantErr:  true,
			expected: 10,
		},
	}

	for _, tt := range tests {
		m := newMetrics()
		for _, v := range tt.values {
			if err := m.processCounter([]byte("counter1"), []byte(v)); err != nil {
				assert.True(t, tt.wantErr, "processCounter(%q) got an unexpected error: %v",
					v, err)
			}
		}
		got := m.Counters()["counter1"]
		assert.Equal(t, tt.expected, got, "Counter value mismatch")
	}
}

func TestProcessGauge(t *testing.T) {
	tests := []struct {
		values   []string
		expected int64
		wantErr  bool
	}{
		{
			values:   []string{"1"},
			expected: 1,
		},
		{
			values:   []string{"1", "2", "3"},
			expected: 3,
		},
		{
			values:  []string{"1.0"},
			wantErr: true,
		},
		{
			values:  []string{"abc"},
			wantErr: true,
		},
		{
			values:   []string{"1", "abc", "9"},
			wantErr:  true,
			expected: 9,
		},
	}

	for _, tt := range tests {
		m := newMetrics()
		for _, v := range tt.values {
			if err := m.processGauge([]byte("g1"), []byte(v)); err != nil {
				assert.True(t, tt.wantErr, "processGauge(%q) got an unexpected error: %v",
					v, err)
			}
		}
		got := m.Gauges()["g1"]
		assert.Equal(t, tt.expected, got, "Gauge value mismatch")
	}
}

func TestProcessTimer(t *testing.T) {
	tests := []struct {
		values   []string
		expected []time.Duration
		wantErr  bool
	}{
		{
			values:   []string{"1.5"},
			expected: []time.Duration{1500 * time.Microsecond},
		},
		{
			values:   []string{"1", "2", "3"},
			expected: []time.Duration{time.Millisecond, 2 * time.Millisecond, 3 * time.Millisecond},
		},
		{
			values:   []string{"1"},
			expected: []time.Duration{time.Millisecond},
		},
		{
			values:  []string{"abc"},
			wantErr: true,
		},
		{
			values:   []string{"1.35", "abc", "9.123456789"},
			wantErr:  true,
			expected: []time.Duration{1350 * time.Microsecond, 9123456 * time.Nanosecond},
		},
	}

	for _, tt := range tests {
		m := newMetrics()
		for _, v := range tt.values {
			if err := m.processTimer([]byte("t1"), []byte(v)); err != nil {
				assert.True(t, tt.wantErr, "processTimer(%q) got an unexpected error: %v",
					v, err)
			}
		}
		got := m.Timers()["t1"]
		assert.Equal(t, tt.expected, got, "Timer values mismatch")
	}
}
