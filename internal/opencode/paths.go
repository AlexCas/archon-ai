package opencode

import (
	"os"
	"path/filepath"
)

// SettingsPathFor returns the path to the opencode settings file under homeDir:
// <homeDir>/.config/opencode/opencode.json.
// Returns "" when homeDir is empty.
func SettingsPathFor(homeDir string) string {
	if homeDir == "" {
		return ""
	}
	return filepath.Join(homeDir, ".config", "opencode", "opencode.json")
}

// CachePathFor returns the path to the opencode models cache under homeDir:
// <homeDir>/.cache/opencode/models.json.
// Returns "" when homeDir is empty.
func CachePathFor(homeDir string) string {
	if homeDir == "" {
		return ""
	}
	return filepath.Join(homeDir, ".cache", "opencode", "models.json")
}

// SettingsPath returns the default path to the opencode settings file using
// the current user's home directory: ~/.config/opencode/opencode.json.
// Returns "" when the home directory is unavailable.
//
// Prefer SettingsPathFor(homeDir) when a home directory is already known,
// so callers (e.g. init, tests) remain hermetic.
func SettingsPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return SettingsPathFor(home)
}

// CachePath returns the default path to the opencode models cache using the
// current user's home directory: ~/.cache/opencode/models.json.
// Returns "" when the home directory is unavailable.
//
// Prefer CachePathFor(homeDir) when a home directory is already known,
// so callers (e.g. init, tests) remain hermetic.
func CachePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return CachePathFor(home)
}
