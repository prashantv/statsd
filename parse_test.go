package statsd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
