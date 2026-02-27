package app

import (
	"testing"

	"github.com/VectorBlue-06/open-llama/internal/config"
	"github.com/VectorBlue-06/open-llama/internal/utils"
)

func TestNewApp(t *testing.T) {
	cfg := config.Defaults()
	logger, _ := utils.NewLogger("/tmp/test-openllama.log", false)
	defer logger.Close()

	a := New(cfg, logger)
	if a == nil {
		t.Fatal("expected non-nil app")
	}
}

func TestAppShutdown(t *testing.T) {
	cfg := config.Defaults()
	logger, _ := utils.NewLogger("/tmp/test-openllama.log", false)

	a := New(cfg, logger)
	// Shutdown should not panic
	a.Shutdown()
}
