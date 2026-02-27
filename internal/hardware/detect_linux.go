//go:build linux

package hardware

import (
	"os/exec"
	"strconv"
	"strings"
)

func detectRAM(info *HardwareInfo) {
	// Use /proc/meminfo for RAM detection on Linux
	data, err := exec.Command("cat", "/proc/meminfo").Output()
	if err != nil {
		return
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "MemTotal:") {
			info.TotalRAM = parseMemInfoValue(line)
		} else if strings.HasPrefix(line, "MemAvailable:") {
			info.FreeRAM = parseMemInfoValue(line)
		}
	}
}

func parseMemInfoValue(line string) uint64 {
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return 0
	}
	val, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return 0
	}
	return val * 1024 // meminfo values are in kB
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
	// Parse: "NVIDIA GeForce RTX 4090, 24564 MiB, 535.129.03"
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
