package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/VectorBlue-06/open-llama/internal/app"
	"github.com/VectorBlue-06/open-llama/internal/config"
	"github.com/VectorBlue-06/open-llama/internal/utils"
)

var version = "dev"

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "OpenLlama crashed: %v\n%s\n", r, debug.Stack())
			fmt.Fprintf(os.Stderr, "See ~/.openllama/openllama.log for details.\n")
			os.Exit(1)
		}
	}()

	// CLI flags
	debugFlag := flag.Bool("debug", false, "Enable debug logging")
	configPath := flag.String("config", "", "Path to config file")
	modelFlag := flag.String("model", "", "Model filename to use")
	portFlag := flag.Int("port", 0, "Port for llama-server (0 = auto)")
	versionFlag := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("OpenLlama %s\n", version)
		os.Exit(0)
	}

	// Load config
	var cfg *config.Config
	var err error
	if *configPath != "" {
		cfg, err = config.LoadFrom(*configPath)
	} else {
		cfg, err = config.Load()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Apply CLI overrides
	if *debugFlag {
		cfg.Debug = true
	}
	if *modelFlag != "" {
		cfg.Model.Default = *modelFlag
	}
	if *portFlag != 0 {
		cfg.Server.Port = *portFlag
	}

	// Initialize logger
	logPath, err := config.LogPath()
	if err != nil {
		logPath = "/tmp/openllama.log"
	}
	logger, err := utils.NewLogger(logPath, cfg.Debug)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not initialize logger: %v\n", err)
	}

	// Run application
	application := app.New(cfg, logger)
	defer application.Shutdown()

	if err := application.Run(); err != nil {
		logger.Error("application error: %v", err)
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
