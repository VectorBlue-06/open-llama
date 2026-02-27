//go:build windows

package hardware

import (
	"os/exec"
	"strconv"
	"strings"
)

func detectRAM(info *HardwareInfo) {
	// Use wmic for Windows RAM detection
	cmd := exec.Command("wmic", "OS", "get", "FreePhysicalMemory,TotalVisibleMemorySize", "/Value")
	output, err := cmd.Output()
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "TotalVisibleMemorySize=") {
			valStr := strings.TrimPrefix(line, "TotalVisibleMemorySize=")
			if val, err := strconv.ParseUint(strings.TrimSpace(valStr), 10, 64); err == nil {
				info.TotalRAM = val * 1024 // wmic values are in kB
			}
		} else if strings.HasPrefix(line, "FreePhysicalMemory=") {
			valStr := strings.TrimPrefix(line, "FreePhysicalMemory=")
			if val, err := strconv.ParseUint(strings.TrimSpace(valStr), 10, 64); err == nil {
				info.FreeRAM = val * 1024 // wmic values are in kB
			}
		}
	}
}

func detectGPU(info *HardwareInfo) {
	detectCUDA(info)
}

func detectCUDA(info *HardwareInfo) {
	cmd := exec.Command("nvidia-smi",
		"--query-gpu=name,memory.total,driver_version",
		"--format=csv,noheader")
	output, err := cmd.Output()
	if err != nil {
		return
	}
	line := strings.TrimSpace(string(output))
	if line == "" {
		return
	}
	parts := strings.SplitN(line, ", ", 3)
	info.HasCUDA = true
	if len(parts) >= 1 {
		info.GPUName = strings.TrimSpace(parts[0])
	}
	if len(parts) >= 2 {
		vramStr := strings.TrimSuffix(strings.TrimSpace(parts[1]), " MiB")
		if vram, err := strconv.ParseUint(vramStr, 10, 64); err == nil {
			info.GPUVRAM = vram * 1024 * 1024
		}
	}
	if len(parts) >= 3 {
		info.CUDAVersion = strings.TrimSpace(parts[2])
	}
}
