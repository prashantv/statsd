package main

import (
	"encoding/json"
	"errors"
	"time"
)

type durationMS time.Duration

func (d durationMS) MarshalJSON() ([]byte, error) {
	ms := float64(time.Duration(d)) / float64(time.Millisecond)
	return json.Marshal(ms)
}

func (d *durationMS) UnmarshalJSON(bs []byte) error {
	return errors.New("cannot UnmarshalJSON durationMS")
}

type byDuration []time.Duration

func (p byDuration) Len() int           { return len(p) }
func (p byDuration) Less(i, j int) bool { return p[i] < p[j] }
func (p byDuration) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
