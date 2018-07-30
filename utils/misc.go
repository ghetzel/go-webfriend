package utils

import "time"

func FudgeDuration(duration time.Duration) time.Duration {
	// allow specifying times as integers representing milliseconds
	// by assuming that if you see a timeout less than 1ms, then it was
	// actually specified as an integer and thus came in as an unreasonably
	// small time.Duration
	if duration > 0 && duration < time.Millisecond {
		return time.Duration(int(duration)) * time.Millisecond
	}

	return duration
}
