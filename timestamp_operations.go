package filesystemsearch

import (
	"strings"
	"time"
)

// function to extract the timestamp from a log line
func ExtractTimestamp(logLine string) (time.Time, error) {
	// Split the log line by spaces
	parts := strings.Split(logLine, " ")

	// Extract the timestamp from the first part of the log line
	timestampStr := parts[0] + "T" + parts[1]

	// Parse the timestamp using the correct format string
	timestamp, err := time.Parse("2006-01-02T15:04:05,000", timestampStr)
	if err != nil {
		return time.Time{}, err
	}

	return timestamp, nil
}
