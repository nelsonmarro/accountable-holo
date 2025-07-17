package validator

import "time"

// IsDateAfter checks if the date d is after the specified afterDate.
func IsDateAfter(d time.Time, afterDate time.Time) bool {
	// truncate the time part of both dates to compare only the date part
	dDateOnly := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
	afterDateOnly := time.Date(afterDate.Year(), afterDate.Month(), afterDate.Day(), 0, 0, 0, 0, time.UTC)

	return dDateOnly.After(afterDateOnly)
}
