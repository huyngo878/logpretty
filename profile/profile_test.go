package profile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setTempHome sets both HOME and USERPROFILE so os.UserHomeDir() works
// on both Linux (CI) and Windows (local).
func setTempHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
	return dir
}

func TestLoad_NoFile(t *testing.T) {
	setTempHome(t)
	cfg, err := Load()
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Empty(t, cfg.LicenseKey)
	assert.NotNil(t, cfg.Profiles)
}

func TestLoad_ValidFile(t *testing.T) {
	dir := setTempHome(t)
	cfgDir := filepath.Join(dir, ".logpretty")
	require.NoError(t, os.MkdirAll(cfgDir, 0700))

	content := "license_key: test-key-123\nprofiles:\n  production:\n    level: WARN\n    filter: payment\n"
	require.NoError(t, os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(content), 0600))

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "test-key-123", cfg.LicenseKey)
	assert.Contains(t, cfg.Profiles, "production")
	assert.Equal(t, "WARN", cfg.Profiles["production"].Level)
}

func TestLoad_MalformedFile(t *testing.T) {
	dir := setTempHome(t)
	cfgDir := filepath.Join(dir, ".logpretty")
	require.NoError(t, os.MkdirAll(cfgDir, 0700))
	// Unclosed flow mapping — genuinely invalid YAML
	require.NoError(t, os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte("key: {unclosed bracket"), 0600))

	cfg, err := Load()
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestValidLicense(t *testing.T) {
	assert.False(t, validLicense(""), "empty key should be invalid")
	assert.True(t, validLicense("any-non-empty-key"), "non-empty key should be valid")
}

func TestGet_NoLicense(t *testing.T) {
	setTempHome(t)
	p, err := Get("production")
	require.NoError(t, err)
	assert.Nil(t, p, "should return nil when no license key")
}

func TestGet_WithLicenseProfileNotFound(t *testing.T) {
	dir := setTempHome(t)
	cfgDir := filepath.Join(dir, ".logpretty")
	require.NoError(t, os.MkdirAll(cfgDir, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte("license_key: valid-key\n"), 0600))

	p, err := Get("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, p)
}

func TestGet_WithLicenseAndProfile(t *testing.T) {
	dir := setTempHome(t)
	cfgDir := filepath.Join(dir, ".logpretty")
	require.NoError(t, os.MkdirAll(cfgDir, 0700))

	content := "license_key: valid-key\nprofiles:\n  staging:\n    level: DEBUG\n"
	require.NoError(t, os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(content), 0600))

	p, err := Get("staging")
	require.NoError(t, err)
	require.NotNil(t, p)
	assert.Equal(t, "DEBUG", p.Level)
}
