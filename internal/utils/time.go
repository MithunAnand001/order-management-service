package utils

import "time"

// Now returns the current local time.
func Now() time.Time {
	return time.Now()
}

// FormatRFC3339 formats a time into RFC3339 string.
func FormatRFC3339(t time.Time) string {
	return t.Format(time.RFC3339)
}

// DiffInMinutes returns the difference between two times in minutes.
func DiffInMinutes(t1, t2 time.Time) int {
	return int(t1.Sub(t2).Minutes())
}

// IsExpired checks if the given time has passed.
func IsExpired(t time.Time) bool {
	return Now().After(t)
}
