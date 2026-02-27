package utils

import (
	"os"
	"path/filepath"
)

// EnsureDir creates a directory and all parents if they don't exist.
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0700)
}

// FileExists returns true if the path exists and is a regular file.
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// DirExists returns true if the path exists and is a directory.
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// ListFiles returns all files in a directory matching the given extension.
func ListFiles(dir string, ext string) ([]string, error) {
	var files []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if ext == "" || filepath.Ext(entry.Name()) == ext {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}
	return files, nil
}

// TempDir returns the application's temporary directory path.
func TempDir(appDir string) string {
	return filepath.Join(appDir, "tmp")
}
