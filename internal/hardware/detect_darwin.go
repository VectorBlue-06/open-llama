//go:build darwin

package hardware

import (
	"os/exec"
	"strconv"
	"strings"
)

func detectRAM(info *HardwareInfo) {
	// Use sysctl for macOS RAM detection
	output, err := exec.Command("sysctl", "-n", "hw.memsize").Output()
	if err != nil {
		return
	}
	val, err := strconv.ParseUint(strings.TrimSpace(string(output)), 10, 64)
	if err != nil {
		return
	}
	info.TotalRAM = val
	// Approximate free RAM from vm_stat
	info.FreeRAM = val / 2 // rough estimate; precise requires vm_stat parsing
}

func detectGPU(info *HardwareInfo) {
	detectMetal(info)
}

func detectMetal(info *HardwareInfo) {
	cmd := exec.Command("system_profiler", "SPDisplaysDataType")
	output, err := cmd.Output()
	if err != nil {
		return
	}
	content := string(output)
	if strings.Contains(content, "Metal") {
		info.HasMetal = true
	}
	// Try to extract GPU name
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "Chipset Model:") {
			info.GPUName = strings.TrimSpace(strings.TrimPrefix(trimmed, "Chipset Model:"))
		}
	}
}
