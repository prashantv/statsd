package statsd

import "time"

func copyMapInt64(m map[string]int64) map[string]int64 {
	newMap := make(map[string]int64)
	for k, v := range m {
		newMap[k] = v
	}
	return newMap
}

func copyMapDuration(m map[string][]time.Duration) map[string][]time.Duration {
	newMap := make(map[string][]time.Duration)
	for k, v := range m {
		newMap[k] = v
	}
	return newMap
}
