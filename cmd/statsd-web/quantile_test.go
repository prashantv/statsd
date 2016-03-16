package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetQuantile(t *testing.T) {
	seq10 := []time.Duration{0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
	tests := []struct {
		times    []time.Duration
		quantile float64
		want     time.Duration
	}{
		{nil, 0.0, 0},
		{nil, 0.5, 0},
		{nil, 1.0, 0},
		{nil, 1.0, 0},
		{[]time.Duration{1}, 0.0, 1},
		{[]time.Duration{1}, 0.5, 1},
		{[]time.Duration{1}, 1.0, 1},
		{seq10, 0.0, 0},
		{seq10, 0.5, 50},
		{seq10, 1.0, 100},
		{seq10, 0.2, 20},
		{seq10, 0.3, 30},
		{seq10, 0.25, 25},
	}

	for _, tt := range tests {
		got := getQuantile(tt.times, tt.quantile)
		assert.Equal(t, tt.want, got, "P%v of %v mismatch", tt.quantile, tt.times)
	}
}
