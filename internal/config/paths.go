package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const appDirName = ".openllama"
const legacyModelsDir = "~/.openllama/models"

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

// RuntimeDirs returns runtime directories near the openllama executable.
// Search order:
// 1. <exe_dir>/runtime
// 2. <exe_dir>/../runtime (useful when binary is in ./bin)
func RuntimeDirs() []string {
	exePath, err := os.Executable()
	if err != nil {
		return []string{filepath.Join("runtime")}
	}

	exeDir := filepath.Dir(exePath)
	parentDir := filepath.Dir(exeDir)
	if strings.EqualFold(filepath.Base(exeDir), "bin") {
		return []string{filepath.Join(parentDir, "runtime"), filepath.Join(exeDir, "runtime")}
	}

	dirs := []string{filepath.Join(exeDir, "runtime")}
	parentRuntime := filepath.Join(parentDir, "runtime")
	if parentRuntime != dirs[0] {
		dirs = append(dirs, parentRuntime)
	}
	return dirs
}

// PrimaryRuntimeDir returns the preferred runtime directory.
func PrimaryRuntimeDir() string {
	return RuntimeDirs()[0]
}

// ResolveModelsDir returns the effective models directory.
// If modelsDir is empty or legacy default, runtime/models near the executable is preferred.
func ResolveModelsDir(modelsDir string) string {
	normalized := strings.TrimSpace(modelsDir)
	if normalized == "" || normalized == legacyModelsDir {
		for _, runtimeDir := range RuntimeDirs() {
			candidate := filepath.Join(runtimeDir, "models")
			if info, err := os.Stat(candidate); err == nil && info.IsDir() {
				return candidate
			}
		}
		return filepath.Join(RuntimeDirs()[0], "models")
	}

	return ExpandPath(normalized)
}

// EnsureDirectories creates all necessary application directories.
func EnsureDirectories(cfg *Config) error {
	modelsDir := ResolveModelsDir(cfg.Model.ModelsDir)
	runtimeDir := PrimaryRuntimeDir()
	dirs := []string{
		ExpandPath("~/.openllama"),
		runtimeDir,
		filepath.Join(runtimeDir, "llama.cpp"),
		modelsDir,
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
