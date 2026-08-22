package report

import (
	"errors"
	"fmt"
	"time"
)

func parseDateRange(startValue, endValue string, required, includeNanoseconds bool) (*time.Time, *time.Time, error) {
	if required && (startValue == "" || endValue == "") {
		return nil, nil, errors.New("statistics date start and end are required")
	}

	var start, end *time.Time
	if startValue != "" {
		parsed, err := time.Parse("2006-01-02", startValue)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid date_start: %w", err)
		}
		start = &parsed
	}
	if endValue != "" {
		parsed, err := time.Parse("2006-01-02", endValue)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid date_end: %w", err)
		}
		parsed = parsed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		if includeNanoseconds {
			parsed = parsed.Add(999999999 * time.Nanosecond)
		}
		end = &parsed
	}
	if start != nil && end != nil && start.After(*end) {
		return nil, nil, errors.New("date start is after end")
	}
	return start, end, nil
}
