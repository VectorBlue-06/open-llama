package models

import "fmt"

// ModelInfo holds information about a discovered GGUF model.
type ModelInfo struct {
	Filename       string `json:"filename"`
	FilePath       string `json:"file_path"`
	FileSize       int64  `json:"file_size"`
	QuantType      string `json:"quant_type"`
	ParameterCount string `json:"parameter_count"`
	Architecture   string `json:"architecture"`
	ContextLength  int    `json:"context_length"`
	RAMEstimate    uint64 `json:"ram_estimate"`
}

// EstimateRAM estimates the RAM needed for a model based on file size.
// Uses a 1.2x multiplier to account for KV cache and runtime overhead.
func EstimateRAM(fileSize int64) uint64 {
	return uint64(float64(fileSize) * 1.2)
}

// FormatSize returns a human-readable file size string.
func FormatSize(bytes int64) string {
	const (
		MB = 1024 * 1024
		GB = 1024 * 1024 * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
