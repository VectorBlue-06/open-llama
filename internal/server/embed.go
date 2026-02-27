package server

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// FindBinary locates the llama-server binary by checking several locations.
// Search order:
// 1. Same directory as the running executable (sidecar mode)
// 2. ~/.openllama/bin/
// 3. PATH lookup
func FindBinary(appDir string) (string, error) {
	binaryName := "llama-server"
	if runtime.GOOS == "windows" {
		binaryName = "llama-server.exe"
	}

	// 1. Same directory as the running executable
	exePath, err := os.Executable()
	if err == nil {
		sidecar := filepath.Join(filepath.Dir(exePath), binaryName)
		if isExecutable(sidecar) {
			return sidecar, nil
		}
	}

	// 2. ~/.openllama/bin/
	binPath := filepath.Join(appDir, "bin", binaryName)
	if isExecutable(binPath) {
		return binPath, nil
	}

	// 3. PATH lookup
	pathBinary, err := exec.LookPath(binaryName)
	if err == nil {
		return pathBinary, nil
	}

	return "", fmt.Errorf("llama-server binary not found; place it next to the openllama binary, in %s/bin/, or in PATH", appDir)
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if info.IsDir() {
		return false
	}
	// On Unix, check execute permission
	if runtime.GOOS != "windows" {
		return info.Mode()&0111 != 0
	}
	return true
}
