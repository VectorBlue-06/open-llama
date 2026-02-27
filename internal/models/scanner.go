package models

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ScanModels scans a directory for GGUF model files and returns info for each.
func ScanModels(dir string) ([]ModelInfo, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var models []ModelInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(strings.ToLower(entry.Name()), ".gguf") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		model := ModelInfo{
			Filename:       entry.Name(),
			FilePath:       filepath.Join(dir, entry.Name()),
			FileSize:       info.Size(),
			QuantType:      parseQuantType(entry.Name()),
			ParameterCount: parseParameterCount(entry.Name()),
			RAMEstimate:    EstimateRAM(info.Size()),
		}
		models = append(models, model)
	}

	return models, nil
}

// parseQuantType extracts the quantization type from a filename.
// e.g., "mistral-7b-q4_k_m.gguf" -> "Q4_K_M"
func parseQuantType(filename string) string {
	re := regexp.MustCompile(`(?i)(q[0-9]+_[a-z0-9_]+)`)
	match := re.FindString(filename)
	if match != "" {
		return strings.ToUpper(match)
	}
	// Try simpler patterns like Q4, Q5, Q8
	re2 := regexp.MustCompile(`(?i)(q[0-9]+)`)
	match = re2.FindString(filename)
	if match != "" {
		return strings.ToUpper(match)
	}
	return "unknown"
}

// parseParameterCount extracts parameter count from filename.
// e.g., "mistral-7b-q4_k_m.gguf" -> "7B"
func parseParameterCount(filename string) string {
	re := regexp.MustCompile(`(?i)(\d+\.?\d*)[bB]`)
	match := re.FindStringSubmatch(filename)
	if len(match) >= 2 {
		return strings.ToUpper(match[1] + "B")
	}
	return "unknown"
}
