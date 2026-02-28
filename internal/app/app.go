package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

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
	cfg               *config.Config
	logger            *utils.Logger
	hw                *hardware.HardwareInfo
	server            *server.Server
	engine            *chat.Engine
	tmplEng           *templates.Engine
	ctxMgr            *ctx.Manager
	metrics           *metrics.Collector
	uiModel           ui.Model
	program           *tea.Program
	modelList         []models.ModelInfo
	currentModel      string
	llamaPathOverride string
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
	modelsDir := config.ResolveModelsDir(a.cfg.Model.ModelsDir)
	a.modelList, err = models.ScanModels(modelsDir)
	if err != nil {
		a.logger.Warn("model scan failed: %v", err)
	}
	if len(a.modelList) == 0 {
		legacyModelsDir := config.ExpandPath("~/.openllama/models")
		if legacyModelsDir != modelsDir {
			legacyModels, legacyErr := models.ScanModels(legacyModelsDir)
			if legacyErr == nil && len(legacyModels) > 0 {
				a.logger.Info("using legacy models directory: %s", legacyModelsDir)
				a.modelList = legacyModels
			}
		}
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
	a.uiModel.SetStartupInfo(hw.TotalRAM, hw.FreeRAM, hw.HasCUDA || hw.HasMetal)
	a.uiModel.SetGenerationSettings(
		a.cfg.Generation.Temperature,
		a.cfg.Generation.TopP,
		a.cfg.Generation.TopK,
		a.cfg.Generation.RepeatPenalty,
		a.cfg.Generation.MaxTokens,
	)
	a.uiModel.SetFontSize(a.cfg.UI.FontSize)
	llamaPath := strings.TrimSpace(a.cfg.Model.LlamaPath)
	if llamaPath == "" {
		llamaPath = filepath.Join(config.PrimaryRuntimeDir(), "llama.cpp")
	}
	a.llamaPathOverride = llamaPath
	a.uiModel.SetRuntimePaths(llamaPath, modelsDir)
	a.uiModel.SetCallbacks(
		a.handleSend,
		a.handleQuit,
		a.handleNewChat,
		a.handleModelSwitch,
	)
	a.uiModel.SetApplySettingsCallback(a.handleApplySettings)

	if len(a.modelList) == 0 {
		a.uiModel.SetError(fmt.Errorf("no models found in %s", modelsDir))
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
	binaryPath, err := a.resolveLlamaBinary()
	if err != nil {
		a.logger.Error("server binary not found: %v", err)
		a.uiModel.SetError(err)
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

	serverLogPath := a.serverLogPath()

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

	// 10. Launch UI (startup screen remains visible; readiness is handled in background)

	return a.runUI()
}

func (a *App) runUI() error {
	a.program = tea.NewProgram(a.uiModel, tea.WithAltScreen())

	if a.server != nil && a.server.IsRunning() {
		go func() {
			if err := a.server.WaitForReady(120 * time.Second); err != nil {
				a.sendUI(ui.ServerFailedMsg{Err: err})
				return
			}
			a.sendUI(ui.ServerReadyMsg{})
		}()
	}

	_, err := a.program.Run()
	a.program = nil
	return err
}

func (a *App) serverLogPath() string {
	appDir, _ := config.AppDir()
	if appDir == "" {
		return ""
	}
	return filepath.Join(appDir, "llama-server.log")
}

func (a *App) resolveLlamaBinary() (string, error) {
	if a.llamaPathOverride != "" {
		binaryName := "llama-server"
		if runtime.GOOS == "windows" {
			binaryName = "llama-server.exe"
		}

		path := a.llamaPathOverride
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			path = filepath.Join(path, binaryName)
		}
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
		return "", fmt.Errorf("llama.cpp not found at %s", path)
	}

	appDir, _ := config.AppDir()
	return server.FindBinary(appDir)
}

func (a *App) waitServerReadyAsync() {
	go func() {
		if err := a.server.WaitForReady(120 * time.Second); err != nil {
			a.sendUI(ui.ServerFailedMsg{Err: err})
			return
		}
		a.sendUI(ui.ServerReadyMsg{})
	}()
}

func (a *App) restartServerForModel(model models.ModelInfo) error {
	if a.server != nil && a.server.IsRunning() {
		if err := a.server.Stop(); err != nil {
			a.logger.Warn("server stop before restart failed: %v", err)
		}
	}

	binaryPath, err := a.resolveLlamaBinary()
	if err != nil {
		return err
	}

	threads := a.cfg.Server.Threads
	if threads == 0 {
		threads = hardware.RecommendThreads(a.hw.CPUCores)
	}
	gpuLayers := a.cfg.Server.GPULayers
	if gpuLayers == -1 {
		gpuLayers = hardware.RecommendGPULayers(a.hw)
	}
	ctxSize := hardware.RecommendCtxSize(a.hw.FreeRAM, a.cfg.Server.CtxSize)

	err = a.server.Start(server.Config{
		BinaryPath: binaryPath,
		ModelPath:  model.FilePath,
		Host:       a.cfg.Server.Host,
		Port:       a.cfg.Server.Port,
		CtxSize:    ctxSize,
		Threads:    threads,
		GPULayers:  gpuLayers,
		BatchSize:  a.cfg.Server.BatchSize,
		ExtraArgs:  a.cfg.Server.ExtraArgs,
	}, a.serverLogPath())
	if err != nil {
		return err
	}

	chatCfg := chat.EngineConfig{
		Temperature:   a.cfg.Generation.Temperature,
		TopP:          a.cfg.Generation.TopP,
		TopK:          a.cfg.Generation.TopK,
		RepeatPenalty: a.cfg.Generation.RepeatPenalty,
		MaxTokens:     a.cfg.Generation.MaxTokens,
	}
	a.engine = chat.NewEngine(a.server.Port(), chatCfg, a.tmplEng)

	a.currentModel = model.Filename
	a.uiModel.SetModelInfo(model.Filename, a.tmplEng.Template().Name)
	a.uiModel.ShowLoading()
	a.waitServerReadyAsync()
	return nil
}

func (a *App) sendUI(msg tea.Msg) {
	if a.program != nil {
		a.program.Send(msg)
	}
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
		a.sendUI(ui.StreamErrorMsg{Err: fmt.Errorf("chat engine not initialized")})
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
		a.sendUI(ui.StreamErrorMsg{Err: err})
		return
	}

	// Stream tokens
	go func() {
		var fullContent string
		var timings *chat.Timings
		for token := range ch {
			if token.Content != "" {
				a.sendUI(ui.StreamChunkMsg{Content: token.Content})
				fullContent += token.Content
			}
			if token.Timings != nil {
				timings = token.Timings
			}
		}

		a.sendUI(ui.StreamDoneMsg{
			FullContent: fullContent,
			Timings:     timings,
		})

		if fullContent == "" {
			return
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
	if err := a.restartServerForModel(a.modelList[index]); err != nil {
		a.logger.Error("model switch failed: %v", err)
		a.sendUI(ui.StreamErrorMsg{Err: err})
	}
}

func (a *App) handleApplySettings(s ui.SettingsUpdate) {
	resolvedModelsPath := config.ResolveModelsDir(s.ModelsPath)
	list, err := models.ScanModels(resolvedModelsPath)
	if err != nil {
		a.sendUI(ui.StreamErrorMsg{Err: err})
		return
	}
	if len(list) == 0 {
		a.sendUI(ui.StreamErrorMsg{Err: fmt.Errorf("no models found in %s", resolvedModelsPath)})
		return
	}

	a.modelList = list
	a.uiModel.SetModels(list)

	idx := s.SelectedModel
	if idx < 0 || idx >= len(list) {
		idx = 0
	}

	llamaPath := strings.TrimSpace(s.LlamaPath)
	if llamaPath == "" {
		llamaPath = filepath.Join(config.PrimaryRuntimeDir(), "llama.cpp")
	}
	a.llamaPathOverride = llamaPath
	a.uiModel.SetRuntimePaths(llamaPath, resolvedModelsPath)

	a.cfg.Model.ModelsDir = s.ModelsPath
	a.cfg.Model.LlamaPath = llamaPath
	a.cfg.Model.Default = list[idx].Filename
	a.cfg.Generation.Temperature = s.Temperature
	a.cfg.Generation.TopP = s.TopP
	a.cfg.Generation.TopK = s.TopK
	a.cfg.Generation.RepeatPenalty = s.RepeatPenalty
	a.cfg.Generation.MaxTokens = s.MaxTokens

	if err := a.restartServerForModel(list[idx]); err != nil {
		a.sendUI(ui.StreamErrorMsg{Err: err})
		return
	}

	if err := config.Save(a.cfg); err != nil {
		a.logger.Warn("save config failed: %v", err)
		a.sendUI(ui.StreamErrorMsg{Err: fmt.Errorf("save config failed: %w", err)})
	}
}

// Shutdown performs cleanup.
func (a *App) Shutdown() {
	a.handleQuit()
	if a.logger != nil {
		a.logger.Close()
	}
}
