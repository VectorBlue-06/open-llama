package hardware

import (
	"testing"
)

func TestRecommendThreads(t *testing.T) {
	tests := []struct {
		cores    int
		expected int
	}{
		{0, 1},
		{1, 1},
		{4, 4},
		{8, 8},
		{16, 8},
		{32, 8},
	}
	for _, tt := range tests {
		got := RecommendThreads(tt.cores)
		if got != tt.expected {
			t.Errorf("RecommendThreads(%d) = %d, want %d", tt.cores, got, tt.expected)
		}
	}
}

func TestRecommendGPULayers(t *testing.T) {
	noGPU := &HardwareInfo{}
	if RecommendGPULayers(noGPU) != 0 {
		t.Error("expected 0 GPU layers with no GPU")
	}

	cuda := &HardwareInfo{HasCUDA: true}
	if RecommendGPULayers(cuda) != 999 {
		t.Error("expected 999 GPU layers with CUDA")
	}

	metal := &HardwareInfo{HasMetal: true}
	if RecommendGPULayers(metal) != 999 {
		t.Error("expected 999 GPU layers with Metal")
	}
}

func TestRecommendCtxSize(t *testing.T) {
	// Less than 4 GB free
	got := RecommendCtxSize(3*1024*1024*1024, 4096)
	if got != 2048 {
		t.Errorf("expected 2048 for low RAM, got %d", got)
	}

	// More than 16 GB free
	got = RecommendCtxSize(20*1024*1024*1024, 4096)
	if got != 8192 {
		t.Errorf("expected 8192 for high RAM, got %d", got)
	}
}

func TestDetect(t *testing.T) {
	info, err := Detect()
	if err != nil {
		t.Fatalf("Detect() failed: %v", err)
	}
	if info.CPUCores <= 0 {
		t.Error("expected CPUCores > 0")
	}
}
