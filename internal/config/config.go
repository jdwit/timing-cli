package config

import (
	"os"
	"os/exec"
	"strings"
	"time"
)

// GetAPIKey returns the API key from TIMING_API_KEY environment variable.
func GetAPIKey() string {
	return os.Getenv("TIMING_API_KEY")
}

// GetTimezone returns the timezone in order of precedence:
// 1. TIMING_TIMEZONE environment variable
// 2. System IANA timezone (resolved from the local clock)
func GetTimezone() string {
	if tz := os.Getenv("TIMING_TIMEZONE"); tz != "" {
		return tz
	}
	return systemTimezone()
}

// systemTimezone tries to resolve the IANA timezone name.
// time.Now().Location().String() returns "Local" on many systems,
// so we read the symlink target of /etc/localtime or fall back to
// the TZ env var. Last resort: UTC.
func systemTimezone() string {
	// Go's Location.String() works when TZ is set explicitly
	if loc := time.Now().Location().String(); loc != "Local" && loc != "" {
		return loc
	}

	// macOS/Linux: /etc/localtime is usually a symlink into zoneinfo
	if target, err := os.Readlink("/etc/localtime"); err == nil {
		if idx := strings.Index(target, "zoneinfo/"); idx != -1 {
			return target[idx+len("zoneinfo/"):]
		}
	}

	// macOS: systemsetup -gettimezone
	if out, err := exec.Command("systemsetup", "-gettimezone").Output(); err == nil {
		// Output: "Time Zone: Europe/Amsterdam"
		if parts := strings.SplitN(string(out), ": ", 2); len(parts) == 2 {
			return strings.TrimSpace(parts[1])
		}
	}

	return "UTC"
}
