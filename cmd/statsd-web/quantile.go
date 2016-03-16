package main

import "time"

// getQuantile gets the given quantile from the list of times.
// It uses linear interpolation.
func getQuantile(times []time.Duration, quantile float64) time.Duration {
	if len(times) == 0 {
		return 0
	}
	if len(times) == 1 {
		return times[0]
	}

	exactIdx := quantile * float64(len(times)-1)
	leftIdx := int(exactIdx)
	if leftIdx == len(times)-1 {
		return times[leftIdx]
	}

	rightIdx := leftIdx + 1
	rightBias := exactIdx - float64(leftIdx)
	leftBias := 1 - rightBias

	return time.Duration(float64(times[rightIdx])*rightBias + float64(times[leftIdx])*leftBias)
}
