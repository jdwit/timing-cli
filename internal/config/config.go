package config

import (
	"os"
	"time"
)

// GetAPIKey returns the API key from TIMING_API_KEY environment variable.
func GetAPIKey() string {
	return os.Getenv("TIMING_API_KEY")
}

// GetTimezone returns the timezone in order of precedence:
// 1. TIMING_TIMEZONE environment variable
// 2. System local timezone
func GetTimezone() string {
	if tz := os.Getenv("TIMING_TIMEZONE"); tz != "" {
		return tz
	}
	return time.Now().Location().String()
}
