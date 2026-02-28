package server

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/VectorBlue-06/open-llama/internal/config"
)

// FindBinary locates the llama-server binary by checking several locations.
// Search order:
// 1. runtime/llama.cpp/ next to openllama executable
// 2. runtime/ next to openllama executable
// 3. Same directory as the running executable (sidecar mode)
// 4. ~/.openllama/bin/
// 5. PATH lookup
func FindBinary(appDir string) (string, error) {
	binaryName := "llama-server"
	if runtime.GOOS == "windows" {
		binaryName = "llama-server.exe"
	}

	for _, runtimeDir := range config.RuntimeDirs() {
		llamaCppBinary := filepath.Join(runtimeDir, "llama.cpp", binaryName)
		if isExecutable(llamaCppBinary) {
			return llamaCppBinary, nil
		}

		runtimeBinary := filepath.Join(runtimeDir, binaryName)
		if isExecutable(runtimeBinary) {
			return runtimeBinary, nil
		}
	}

	// 3. Same directory as the running executable
	exePath, err := os.Executable()
	if err == nil {
		sidecar := filepath.Join(filepath.Dir(exePath), binaryName)
		if isExecutable(sidecar) {
			return sidecar, nil
		}
	}

	// 4. ~/.openllama/bin/
	binPath := filepath.Join(appDir, "bin", binaryName)
	if isExecutable(binPath) {
		return binPath, nil
	}

	// 5. PATH lookup
	pathBinary, err := exec.LookPath(binaryName)
	if err == nil {
		return pathBinary, nil
	}

	primaryRuntime := config.PrimaryRuntimeDir()
	return "", fmt.Errorf("llama.cpp not installed: expected %s; place %s there or in %s/bin/ or PATH", filepath.Join(primaryRuntime, "llama.cpp"), binaryName, appDir)
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
