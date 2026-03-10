package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestConfig(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	origDir := configDir
	origFile := configFile
	configDir = dir
	configFile = filepath.Join(dir, "config.yaml")
	t.Cleanup(func() {
		configDir = origDir
		configFile = origFile
	})
}

// --- Load ---

func TestLoad_DefaultsWhenMissing(t *testing.T) {
	setupTestConfig(t)

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "Europe/Amsterdam", cfg.Timezone)
	assert.Empty(t, cfg.APIKey)
}

func TestLoad_FromFile(t *testing.T) {
	setupTestConfig(t)
	require.NoError(t, os.WriteFile(configFile, []byte("api-key: my-key\ntimezone: US/Pacific\n"), 0o600))

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "my-key", cfg.APIKey)
	assert.Equal(t, "US/Pacific", cfg.Timezone)
}

func TestLoad_InvalidYAML(t *testing.T) {
	setupTestConfig(t)
	require.NoError(t, os.WriteFile(configFile, []byte(":::invalid"), 0o600))

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing config")
}

func TestLoad_ReadError(t *testing.T) {
	origDir := configDir
	origFile := configFile
	// Point at a directory instead of file to trigger read error
	dir := t.TempDir()
	configDir = dir
	configFile = dir // dir, not a file
	t.Cleanup(func() {
		configDir = origDir
		configFile = origFile
	})

	_, err := Load()
	require.Error(t, err)
}

// --- Save ---

func TestSave(t *testing.T) {
	setupTestConfig(t)

	cfg := &Config{APIKey: "saved-key", Timezone: "UTC"}
	require.NoError(t, Save(cfg))

	loaded, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "saved-key", loaded.APIKey)
	assert.Equal(t, "UTC", loaded.Timezone)
}

func TestSave_CreatesNestedDirs(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "sub", "config")
	origDir := configDir
	origFile := configFile
	configDir = subdir
	configFile = filepath.Join(subdir, "config.yaml")
	t.Cleanup(func() {
		configDir = origDir
		configFile = origFile
	})

	require.NoError(t, Save(&Config{APIKey: "key", Timezone: "UTC"}))

	_, err := os.Stat(configFile)
	assert.NoError(t, err)
}

func TestSave_FilePermissions(t *testing.T) {
	setupTestConfig(t)
	require.NoError(t, Save(&Config{APIKey: "secret"}))

	info, err := os.Stat(configFile)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
}

// --- Get ---

func TestGet_APIKey(t *testing.T) {
	setupTestConfig(t)
	require.NoError(t, Save(&Config{APIKey: "k", Timezone: "UTC"}))

	val, err := Get("api-key")
	require.NoError(t, err)
	assert.Equal(t, "k", val)
}

func TestGet_Timezone(t *testing.T) {
	setupTestConfig(t)
	require.NoError(t, Save(&Config{Timezone: "Europe/Berlin"}))

	val, err := Get("timezone")
	require.NoError(t, err)
	assert.Equal(t, "Europe/Berlin", val)
}

func TestGet_InvalidKey(t *testing.T) {
	setupTestConfig(t)

	_, err := Get("invalid")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown config key")
}

// --- Set ---

func TestSet_APIKey(t *testing.T) {
	setupTestConfig(t)
	require.NoError(t, Set("api-key", "new-key"))

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "new-key", cfg.APIKey)
}

func TestSet_Timezone(t *testing.T) {
	setupTestConfig(t)
	require.NoError(t, Set("timezone", "Asia/Tokyo"))

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "Asia/Tokyo", cfg.Timezone)
}

func TestSet_PreservesOtherValues(t *testing.T) {
	setupTestConfig(t)
	require.NoError(t, Save(&Config{APIKey: "original", Timezone: "UTC"}))

	require.NoError(t, Set("timezone", "US/Eastern"))

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "original", cfg.APIKey)
	assert.Equal(t, "US/Eastern", cfg.Timezone)
}

func TestSet_InvalidKey(t *testing.T) {
	setupTestConfig(t)

	err := Set("invalid", "val")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown config key")
	assert.Contains(t, err.Error(), "valid keys")
}

// --- GetAPIKey ---

func TestGetAPIKey_FromEnv(t *testing.T) {
	setupTestConfig(t)
	t.Setenv("TIMING_API_KEY", "env-key")

	assert.Equal(t, "env-key", GetAPIKey())
}

func TestGetAPIKey_EnvTakesPrecedence(t *testing.T) {
	setupTestConfig(t)
	require.NoError(t, Save(&Config{APIKey: "config-key"}))
	t.Setenv("TIMING_API_KEY", "env-key")

	assert.Equal(t, "env-key", GetAPIKey())
}

func TestGetAPIKey_FromConfig(t *testing.T) {
	setupTestConfig(t)
	t.Setenv("TIMING_API_KEY", "")
	require.NoError(t, Save(&Config{APIKey: "config-key", Timezone: "UTC"}))

	assert.Equal(t, "config-key", GetAPIKey())
}

func TestGetAPIKey_Empty(t *testing.T) {
	setupTestConfig(t)
	t.Setenv("TIMING_API_KEY", "")

	assert.Empty(t, GetAPIKey())
}

// --- GetTimezone ---

func TestGetTimezone_Default(t *testing.T) {
	setupTestConfig(t)

	assert.Equal(t, "Europe/Amsterdam", GetTimezone())
}

func TestGetTimezone_FromConfig(t *testing.T) {
	setupTestConfig(t)
	require.NoError(t, Save(&Config{Timezone: "US/Eastern"}))

	assert.Equal(t, "US/Eastern", GetTimezone())
}

func TestGetTimezone_EmptyFallsBack(t *testing.T) {
	setupTestConfig(t)
	require.NoError(t, Save(&Config{Timezone: ""}))

	assert.Equal(t, "Europe/Amsterdam", GetTimezone())
}

// --- Dir / File ---

func TestDir(t *testing.T) {
	assert.NotEmpty(t, Dir())
}

func TestFile(t *testing.T) {
	assert.NotEmpty(t, File())
	assert.Contains(t, File(), "config.yaml")
}
