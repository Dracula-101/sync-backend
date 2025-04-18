package utils

import (
	"strings"
	"time"
)

func ParseDuration(durationStr string) time.Duration {
	durationStr = strings.TrimSpace(durationStr)
	if len(durationStr) == 0 {
		return 0
	}

	// Handle days explicitly
	if strings.Contains(durationStr, "d") {
		parts := strings.Split(durationStr, "d")
		daysPart := strings.TrimSpace(parts[0])
		remainingPart := strings.TrimSpace(strings.Join(parts[1:], "d"))

		days, err := time.ParseDuration(daysPart + "h")
		if err != nil {
			panic(err)
		}

		remainingDuration := time.Duration(0)
		if len(remainingPart) > 0 {
			remainingDuration, err = time.ParseDuration(remainingPart)
			if err != nil {
				panic(err)
			}
		}

		return days*24 + remainingDuration
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		panic(err)
	}
	return duration
}
