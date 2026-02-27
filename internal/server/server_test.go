package server

import (
	"testing"
)

func TestFindFreePort(t *testing.T) {
	port, err := FindFreePort()
	if err != nil {
		t.Fatalf("FindFreePort() failed: %v", err)
	}
	if port <= 0 || port > 65535 {
		t.Errorf("invalid port: %d", port)
	}
}

func TestBuildArgs(t *testing.T) {
	cfg := Config{
		ModelPath: "/path/to/model.gguf",
		Host:      "127.0.0.1",
		Port:      8080,
		CtxSize:   4096,
		Threads:   4,
		GPULayers: 0,
		BatchSize: 512,
	}

	args := buildArgs(cfg)

	// Check that essential args are present
	found := map[string]bool{}
	for i, arg := range args {
		if arg == "-m" && i+1 < len(args) {
			found["model"] = args[i+1] == "/path/to/model.gguf"
		}
		if arg == "--port" && i+1 < len(args) {
			found["port"] = args[i+1] == "8080"
		}
		if arg == "--ctx-size" && i+1 < len(args) {
			found["ctx"] = args[i+1] == "4096"
		}
	}

	if !found["model"] {
		t.Error("missing or incorrect model path arg")
	}
	if !found["port"] {
		t.Error("missing or incorrect port arg")
	}
	if !found["ctx"] {
		t.Error("missing or incorrect ctx-size arg")
	}
}

func TestBuildArgsWithGPU(t *testing.T) {
	cfg := Config{
		ModelPath: "/path/to/model.gguf",
		Host:      "127.0.0.1",
		Port:      8080,
		CtxSize:   4096,
		Threads:   4,
		GPULayers: 999,
		BatchSize: 512,
	}

	args := buildArgs(cfg)

	hasGPU := false
	for _, arg := range args {
		if arg == "--n-gpu-layers" {
			hasGPU = true
		}
	}

	if !hasGPU {
		t.Error("expected --n-gpu-layers arg when GPULayers > 0")
	}
}

func TestNewServer(t *testing.T) {
	s := New()
	if s == nil {
		t.Error("expected non-nil server")
	}
	if s.Port() != 0 {
		t.Errorf("expected port 0 before start, got %d", s.Port())
	}
}

func TestServerIsNotRunning(t *testing.T) {
	s := New()
	if s.IsRunning() {
		t.Error("expected server to not be running before start")
	}
}
