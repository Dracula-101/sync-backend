package util

import "time"

func ParseDuration(duration string) time.Duration {
	// Parse the duration string using time.ParseDuration
	d, err := time.ParseDuration(duration)
	if err != nil {
		panic("Invalid duration format: " + duration)
	}
	return d
}
