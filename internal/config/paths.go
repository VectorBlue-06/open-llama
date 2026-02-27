package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const appDirName = ".openllama"

// AppDir returns the base application directory (~/.openllama/).
func AppDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, appDirName), nil
}

// ConfigPath returns the path to the config file.
func ConfigPath() (string, error) {
	dir, err := AppDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// LogPath returns the path to the log file.
func LogPath() (string, error) {
	dir, err := AppDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "openllama.log"), nil
}

// ExpandPath expands ~ to the user's home directory.
func ExpandPath(path string) string {
	if path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return home
	}
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

// EnsureDirectories creates all necessary application directories.
func EnsureDirectories(cfg *Config) error {
	dirs := []string{
		ExpandPath("~/.openllama"),
		ExpandPath(cfg.Model.ModelsDir),
		ExpandPath(cfg.Session.SessionsDir),
		ExpandPath("~/.openllama/tmp"),
		ExpandPath("~/.openllama/bin"),
	}
	perm := os.FileMode(0700)
	if runtime.GOOS == "windows" {
		perm = os.FileMode(0755)
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, perm); err != nil {
			return err
		}
	}
	return nil
}
