package hardware

import (
	"runtime"
)

// HardwareInfo contains detected hardware capabilities.
type HardwareInfo struct {
	CPUCores    int    `json:"cpu_cores"`
	TotalRAM    uint64 `json:"total_ram"`
	FreeRAM     uint64 `json:"free_ram"`
	HasCUDA     bool   `json:"has_cuda"`
	CUDAVersion string `json:"cuda_version,omitempty"`
	HasMetal    bool   `json:"has_metal"`
	GPUName     string `json:"gpu_name,omitempty"`
	GPUVRAM     uint64 `json:"gpu_vram,omitempty"`
}

// Detect detects the hardware capabilities of the current system.
func Detect() (*HardwareInfo, error) {
	info := &HardwareInfo{
		CPUCores: runtime.NumCPU(),
	}

	detectRAM(info)
	detectGPU(info)

	return info, nil
}

// RecommendThreads returns the recommended thread count based on CPU cores.
func RecommendThreads(cores int) int {
	if cores <= 0 {
		return 1
	}
	if cores > 8 {
		return 8
	}
	return cores
}

// RecommendGPULayers returns the recommended GPU layers based on detection.
func RecommendGPULayers(info *HardwareInfo) int {
	if info.HasCUDA || info.HasMetal {
		return 999 // offload all layers
	}
	return 0
}

// RecommendCtxSize returns the recommended context size based on available RAM.
func RecommendCtxSize(freeRAM uint64, configuredSize int) int {
	gb := freeRAM / (1024 * 1024 * 1024)
	maxCtx := configuredSize

	if gb < 4 && maxCtx > 2048 {
		maxCtx = 2048
	} else if gb >= 16 && maxCtx < 8192 {
		maxCtx = 8192
	}

	return maxCtx
}
