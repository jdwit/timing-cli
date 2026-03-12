package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- GetAPIKey ---

func TestGetAPIKey_FromEnv(t *testing.T) {
	t.Setenv("TIMING_API_KEY", "env-key")
	assert.Equal(t, "env-key", GetAPIKey())
}

func TestGetAPIKey_Empty(t *testing.T) {
	t.Setenv("TIMING_API_KEY", "")
	assert.Empty(t, GetAPIKey())
}

// --- GetTimezone ---

func TestGetTimezone_FromEnv(t *testing.T) {
	t.Setenv("TIMING_TIMEZONE", "US/Pacific")
	assert.Equal(t, "US/Pacific", GetTimezone())
}

func TestGetTimezone_FallsBackToSystem(t *testing.T) {
	t.Setenv("TIMING_TIMEZONE", "")
	tz := GetTimezone()
	assert.NotEmpty(t, tz)
}
