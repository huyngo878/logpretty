package profile

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const proMessage = "This feature requires logpretty pro. See https://logpretty.dev"

// Config holds the full user config stored in ~/.logpretty/config.yaml.
type Config struct {
	LicenseKey string             `yaml:"license_key"`
	Profiles   map[string]Profile `yaml:"profiles"`
}

// Profile is a named set of filter options.
type Profile struct {
	Level   string `yaml:"level"`
	Filter  string `yaml:"filter"`
	Since   string `yaml:"since"`
	Format  string `yaml:"format"`
	NoColor bool   `yaml:"no_color"`
}

// Load reads the user config file. Returns an empty config if the file doesn't exist.
func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return &Config{Profiles: make(map[string]Profile)}, nil
	}

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &Config{Profiles: make(map[string]Profile)}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("profile: read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("profile: parse config: %w", err)
	}
	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]Profile)
	}
	return &cfg, nil
}

// Get returns the named profile if it exists and the license is valid.
// Returns nil (not an error) if the license is invalid — callers degrade gracefully.
func Get(name string) (*Profile, error) {
	cfg, err := Load()
	if err != nil {
		return nil, err
	}

	if !validLicense(cfg.LicenseKey) {
		fmt.Fprintln(os.Stderr, proMessage)
		return nil, nil
	}

	p, ok := cfg.Profiles[name]
	if !ok {
		return nil, fmt.Errorf("profile: %q not found", name)
	}
	return &p, nil
}

// validLicense performs a local hash check — no network call.
func validLicense(key string) bool {
	if key == "" {
		return false
	}
	h := sha256.Sum256([]byte(key))
	_ = hex.EncodeToString(h[:])
	// Real validation would compare against a known hash or HMAC.
	// For now, any non-empty key is accepted as a demo.
	return true
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("profile: home dir: %w", err)
	}
	return filepath.Join(home, ".logpretty", "config.yaml"), nil
}
