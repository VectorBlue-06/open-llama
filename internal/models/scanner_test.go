package models

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseQuantType(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"mistral-7b-q4_k_m.gguf", "Q4_K_M"},
		{"llama-3-8b-q5_k_s.gguf", "Q5_K_S"},
		{"phi-3-mini-q8_0.gguf", "Q8_0"},
		{"model.gguf", "unknown"},
	}
	for _, tt := range tests {
		got := parseQuantType(tt.filename)
		if got != tt.expected {
			t.Errorf("parseQuantType(%q) = %q, want %q", tt.filename, got, tt.expected)
		}
	}
}

func TestParseParameterCount(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"mistral-7b-q4.gguf", "7B"},
		{"llama-3-8b.gguf", "8B"},
		{"phi-3-mini-3.8b.gguf", "3.8B"},
		{"model.gguf", "unknown"},
	}
	for _, tt := range tests {
		got := parseParameterCount(tt.filename)
		if got != tt.expected {
			t.Errorf("parseParameterCount(%q) = %q, want %q", tt.filename, got, tt.expected)
		}
	}
}

func TestEstimateRAM(t *testing.T) {
	var fileSize int64 = 4 * 1024 * 1024 * 1024 // 4 GB file
	got := EstimateRAM(fileSize)
	expected := uint64(float64(fileSize) * 1.2)
	if got != expected {
		t.Errorf("EstimateRAM(4GB) = %d, want %d", got, expected)
	}
}

func TestScanModels(t *testing.T) {
	dir := t.TempDir()

	// Create fake GGUF files
	os.WriteFile(filepath.Join(dir, "model-7b-q4_k_m.gguf"), []byte("fake"), 0644)
	os.WriteFile(filepath.Join(dir, "another-13b-q5.gguf"), []byte("data"), 0644)
	os.WriteFile(filepath.Join(dir, "not-a-model.txt"), []byte("txt"), 0644)

	models, err := ScanModels(dir)
	if err != nil {
		t.Fatalf("ScanModels failed: %v", err)
	}
	if len(models) != 2 {
		t.Errorf("expected 2 models, got %d", len(models))
	}
}

func TestScanModelsEmpty(t *testing.T) {
	dir := t.TempDir()
	models, err := ScanModels(dir)
	if err != nil {
		t.Fatalf("ScanModels failed: %v", err)
	}
	if len(models) != 0 {
		t.Errorf("expected 0 models, got %d", len(models))
	}
}

func TestScanModelsNonexistent(t *testing.T) {
	models, err := ScanModels("/nonexistent/path")
	if err != nil {
		t.Fatalf("expected nil error for nonexistent dir, got %v", err)
	}
	if models != nil {
		t.Errorf("expected nil models for nonexistent dir")
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{1073741824, "1.0 GB"},  // 1 GB
		{4831838208, "4.5 GB"},  // 4.5 GB
		{104857600, "100.0 MB"}, // 100 MB
		{500, "500 B"},
	}
	for _, tt := range tests {
		got := FormatSize(tt.bytes)
		if got != tt.expected {
			t.Errorf("FormatSize(%d) = %q, want %q", tt.bytes, got, tt.expected)
		}
	}
}
