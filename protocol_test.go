package statsd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProcessPacket(t *testing.T) {
	empty := newMetrics().FlushAndSnapshot()
	tests := []struct {
		packets  []string
		expected *Snapshot
		wantErr  bool
	}{
		{
			packets: []string{"counter1:1|c"},
			expected: &Snapshot{
				Counters: map[string]int64{
					"counter1": 1,
				},
			},
		},
		{
			packets: []string{"counter1|c:1"},
			wantErr: true,
		},
		{
			packets: []string{"counter1|c"},
			wantErr: true,
		},
		{
			packets: []string{"counter1:1"},
			wantErr: true,
		},
		{
			packets: []string{"counter1:1|"},
			wantErr: true,
		},
		{
			packets: []string{":1|c"},
			wantErr: true,
		},
		{
			packets: []string{"c:|c"},
			wantErr: true,
		},
		{
			packets: []string{"counter1:1|c\ncounter2:2|c\ncounter1:9|c\ncounter1:5|g"},
			expected: &Snapshot{
				Counters: map[string]int64{
					"counter1": 10,
					"counter2": 2,
				},
				Gauges: map[string]int64{
					"counter1": 5,
				},
			},
		},
		{
			packets: []string{"counter1:1|c", "counter1:3|c", "counter1:5|g\nt1:10|ms"},
			expected: &Snapshot{
				Counters: map[string]int64{
					"counter1": 4,
				},
				Gauges: map[string]int64{
					"counter1": 5,
				},
				Timers: map[string][]time.Duration{
					"t1": {10 * time.Millisecond},
				},
			},
		},
	}

	for _, tt := range tests {
		if tt.expected == nil {
			tt.expected = empty
		}
		// Empty metrics are not nil, but are an empty map.
		if tt.expected.Counters == nil {
			tt.expected.Counters = empty.Counters
		}
		if tt.expected.Gauges == nil {
			tt.expected.Gauges = empty.Gauges
		}
		if tt.expected.Timers == nil {
			tt.expected.Timers = empty.Timers
		}

		m := newMetrics()
		for _, packet := range tt.packets {
			if err := m.processPacket([]byte(packet)); err != nil {
				assert.True(t, tt.wantErr, "processPacket(%q) got an unexpected error: %v",
					packet, err)
			}
		}

		got := m.FlushAndSnapshot()
		assert.Equal(t, tt.expected, got, "Snapshot mismatch")
	}
}
