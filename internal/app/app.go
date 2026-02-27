package app

import (
	"context"
	"fmt"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/VectorBlue-06/open-llama/internal/chat"
	"github.com/VectorBlue-06/open-llama/internal/config"
	ctx "github.com/VectorBlue-06/open-llama/internal/context"
	"github.com/VectorBlue-06/open-llama/internal/hardware"
	"github.com/VectorBlue-06/open-llama/internal/metrics"
	"github.com/VectorBlue-06/open-llama/internal/models"
	"github.com/VectorBlue-06/open-llama/internal/server"
	"github.com/VectorBlue-06/open-llama/internal/templates"
	"github.com/VectorBlue-06/open-llama/internal/ui"
	"github.com/VectorBlue-06/open-llama/internal/utils"
)

// App is the main application controller.
type App struct {
	cfg      *config.Config
	logger   *utils.Logger
	hw       *hardware.HardwareInfo
	server   *server.Server
	engine   *chat.Engine
	tmplEng  *templates.Engine
	ctxMgr   *ctx.Manager
	metrics  *metrics.Collector
	uiModel  ui.Model
	modelList []models.ModelInfo
	currentModel string
}

// New creates a new App instance.
func New(cfg *config.Config, logger *utils.Logger) *App {
	return &App{
		cfg:     cfg,
		logger:  logger,
		server:  server.New(),
		metrics: metrics.NewCollector(),
	}
}

// Run starts the application lifecycle.
func (a *App) Run() error {
	a.logger.Info("starting OpenLlama")

	// 1. Ensure directories
	if err := config.EnsureDirectories(a.cfg); err != nil {
		return fmt.Errorf("create directories: %w", err)
	}

	// 2. Detect hardware
	hw, err := hardware.Detect()
	if err != nil {
		a.logger.Warn("hardware detection failed: %v", err)
		hw = &hardware.HardwareInfo{CPUCores: 4}
	}
	a.hw = hw
	a.logger.Info("hardware: %d cores, %d MB RAM, CUDA=%v, Metal=%v",
		hw.CPUCores, hw.TotalRAM/(1024*1024), hw.HasCUDA, hw.HasMetal)

	a.metrics.SetHardware(hw.CPUCores, hw.TotalRAM, hw.HasCUDA || hw.HasMetal,
		hardware.RecommendGPULayers(hw))

	// 3. Scan models
	modelsDir := config.ExpandPath(a.cfg.Model.ModelsDir)
	a.modelList, err = models.ScanModels(modelsDir)
	if err != nil {
		a.logger.Warn("model scan failed: %v", err)
	}
	a.logger.Info("models found: %d", len(a.modelList))

	// 4. Set up template engine
	tmpl := templates.GetByName(a.cfg.Template.Default)
	a.tmplEng = templates.NewEngine(tmpl)

	// 5. Set up context manager
	ctxSize := hardware.RecommendCtxSize(hw.FreeRAM, a.cfg.Server.CtxSize)
	a.ctxMgr = ctx.NewManager(ctxSize, 512)
	a.ctxMgr.SetSystemPrompt(a.cfg.Template.SystemPrompt)

	// 6. Set up UI
	a.uiModel = ui.NewModel(a.metrics)
	a.uiModel.SetModels(a.modelList)
	a.uiModel.SetModelInfo("", a.tmplEng.Template().Name)

	if len(a.modelList) == 0 {
		a.uiModel.ShowWelcome()
		return a.runUI()
	}

	// 7. Select model
	selectedModel := a.selectModel()
	a.currentModel = selectedModel.Filename
	a.uiModel.SetModelInfo(selectedModel.Filename, a.tmplEng.Template().Name)

	// 8. Auto-detect template from filename
	detectedTmpl := templates.DetectFromFilename(selectedModel.Filename)
	a.tmplEng.SetTemplate(detectedTmpl)
	a.uiModel.SetModelInfo(selectedModel.Filename, detectedTmpl.Name)

	// 9. Find and start server
	appDir, _ := config.AppDir()
	binaryPath, err := server.FindBinary(appDir)
	if err != nil {
		a.logger.Error("server binary not found: %v", err)
		// Still launch UI but show error
		return a.runUI()
	}

	threads := a.cfg.Server.Threads
	if threads == 0 {
		threads = hardware.RecommendThreads(hw.CPUCores)
	}
	gpuLayers := a.cfg.Server.GPULayers
	if gpuLayers == -1 {
		gpuLayers = hardware.RecommendGPULayers(hw)
	}

	serverLogPath := ""
	if appDir != "" {
		serverLogPath = filepath.Join(appDir, "llama-server.log")
	}

	srvCfg := server.Config{
		BinaryPath: binaryPath,
		ModelPath:  selectedModel.FilePath,
		Host:       a.cfg.Server.Host,
		Port:       a.cfg.Server.Port,
		CtxSize:    ctxSize,
		Threads:    threads,
		GPULayers:  gpuLayers,
		BatchSize:  a.cfg.Server.BatchSize,
		ExtraArgs:  a.cfg.Server.ExtraArgs,
	}

	a.logger.Info("starting llama-server: model=%s, threads=%d, gpu_layers=%d, ctx=%d",
		selectedModel.Filename, threads, gpuLayers, ctxSize)

	if err := a.server.Start(srvCfg, serverLogPath); err != nil {
		a.logger.Error("server start failed: %v", err)
		return fmt.Errorf("start server: %w", err)
	}

	a.logger.Info("server started on port %d, waiting for ready...", a.server.Port())

	// Set up chat engine
	chatCfg := chat.EngineConfig{
		Temperature:   a.cfg.Generation.Temperature,
		TopP:          a.cfg.Generation.TopP,
		TopK:          a.cfg.Generation.TopK,
		RepeatPenalty: a.cfg.Generation.RepeatPenalty,
		MaxTokens:     a.cfg.Generation.MaxTokens,
	}
	a.engine = chat.NewEngine(a.server.Port(), chatCfg, a.tmplEng)

	// 10. Wait for server ready (in background) and launch UI
	a.uiModel.ShowLoading()

	// Set up callbacks
	a.uiModel.SetCallbacks(
		a.handleSend,
		a.handleQuit,
		a.handleNewChat,
		a.handleModelSwitch,
	)

	return a.runUI()
}

func (a *App) runUI() error {
	p := tea.NewProgram(a.uiModel, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func (a *App) selectModel() models.ModelInfo {
	// If default model is configured, use it
	if a.cfg.Model.Default != "" {
		for _, m := range a.modelList {
			if m.Filename == a.cfg.Model.Default {
				return m
			}
		}
	}
	// If only one model, auto-select
	if len(a.modelList) == 1 {
		return a.modelList[0]
	}
	// Otherwise return first model (UI picker will handle it later)
	return a.modelList[0]
}

func (a *App) handleSend(text string) {
	if a.engine == nil {
		return
	}

	// Add user message to context
	a.ctxMgr.Add(templates.Message{
		Role:    templates.RoleUser,
		Content: text,
	})

	// Build prompt
	prompt := a.ctxMgr.Build(a.tmplEng)

	// Send to server
	bgCtx := context.Background()
	ch, err := a.engine.Send(bgCtx, prompt)
	if err != nil {
		a.logger.Error("send failed: %v", err)
		return
	}

	// Stream tokens
	go func() {
		var fullContent string
		for token := range ch {
			fullContent += token.Content
		}
		// Add assistant message to context
		a.ctxMgr.Add(templates.Message{
			Role:    templates.RoleAssistant,
			Content: fullContent,
		})
	}()
}

func (a *App) handleQuit() {
	a.logger.Info("shutting down")
	if a.server != nil {
		if err := a.server.Stop(); err != nil {
			a.logger.Error("server stop failed: %v", err)
		}
	}
}

func (a *App) handleNewChat() {
	a.ctxMgr.Clear()
	if a.engine != nil {
		a.engine.Reset()
	}
	a.logger.Info("new chat started")
}

func (a *App) handleModelSwitch(index int) {
	if index < 0 || index >= len(a.modelList) {
		return
	}
	a.logger.Info("switching model to %s", a.modelList[index].Filename)
	// Implementation: stop server, restart with new model
}

// Shutdown performs cleanup.
func (a *App) Shutdown() {
	a.handleQuit()
	if a.logger != nil {
		a.logger.Close()
	}
}
