package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/VectorBlue-06/open-llama/internal/metrics"
	"github.com/VectorBlue-06/open-llama/internal/models"
)

// OverlayType represents which overlay is currently shown.
type OverlayType int

const (
	OverlayNone OverlayType = iota
	OverlayModelPicker
	OverlayTemplatePicker
	OverlayWelcome
	OverlayLoading
	OverlaySettings
)

type UIMode int

const (
	ModeStartup UIMode = iota
	ModeChat
)

// ChatMessage represents a displayed chat message.
type ChatMessage struct {
	Role    string
	Content string
}

// SettingsUpdate carries user-edited settings from the settings overlay.
type SettingsUpdate struct {
	LlamaPath     string
	ModelsPath    string
	FontSize      int
	Temperature   float64
	TopP          float64
	TopK          int
	RepeatPenalty float64
	MaxTokens     int
	SelectedModel int
}

// Model is the root Bubble Tea model.
type Model struct {
	// Configuration
	keys         KeyMap
	modelName    string
	templateName string

	// Metrics
	metrics *metrics.Collector

	// UI components
	viewport viewport.Model
	textarea textarea.Model
	spinner  spinner.Model

	// State
	width, height int
	streaming     bool
	mode          UIMode
	showOverlay   OverlayType
	err           error
	messages      []ChatMessage
	streamBuffer  string
	lastRender    time.Time
	ready         bool
	serverReady   bool
	fontSize      int

	// Startup/Settings data
	ramInfo          string
	gpuInfo          string
	llamaPath        string
	modelsPath       string
	temperature      float64
	topP             float64
	topK             int
	repeatPenalty    float64
	maxTokens        int
	settingsInputs   []textinput.Model
	settingsField    int
	settingsEditing  bool
	settingsModelIdx int

	// Models
	availableModels []models.ModelInfo
	selectedModel   int

	// Callbacks
	onSend        func(string)
	onQuit        func()
	onNewChat     func()
	onModelSwitch func(int)
	onApplySettings func(SettingsUpdate)
}

// NewModel creates the root UI model.
func NewModel(metricsCollector *metrics.Collector) Model {
	ta := textarea.New()
	ta.Placeholder = "Type your message..."
	ta.Focus()
	ta.CharLimit = 0
	ta.SetHeight(3)
	ta.ShowLineNumbers = false

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = SpinnerStyle

	return Model{
		keys:          DefaultKeyMap(),
		textarea:      ta,
		spinner:       sp,
		metrics:       metricsCollector,
		messages:      []ChatMessage{},
		lastRender:    time.Now(),
		mode:          ModeStartup,
		fontSize:      2,
		llamaPath:     "runtime/llama.cpp",
		modelsPath:    "runtime/models",
		temperature:   0.7,
		topP:          0.9,
		topK:          40,
		repeatPenalty: 1.1,
		maxTokens:     2048,
	}
}

// SetCallbacks sets the callback functions.
func (m *Model) SetCallbacks(onSend func(string), onQuit func(), onNewChat func(), onModelSwitch func(int)) {
	m.onSend = onSend
	m.onQuit = onQuit
	m.onNewChat = onNewChat
	m.onModelSwitch = onModelSwitch
}

// SetApplySettingsCallback sets the callback for settings updates.
func (m *Model) SetApplySettingsCallback(onApply func(SettingsUpdate)) {
	m.onApplySettings = onApply
}

// SetModelInfo sets the current model display info.
func (m *Model) SetModelInfo(name string, templateName string) {
	m.modelName = name
	m.templateName = templateName
}

// SetStartupInfo sets hardware summary used by startup screen.
func (m *Model) SetStartupInfo(totalRAM, freeRAM uint64, gpuEnabled bool) {
	m.ramInfo = fmt.Sprintf("RAM available: %.1f GB / %.1f GB", float64(freeRAM)/(1024*1024*1024), float64(totalRAM)/(1024*1024*1024))
	if gpuEnabled {
		m.gpuInfo = "GPU: enabled"
	} else {
		m.gpuInfo = "GPU: not detected"
	}
}

// SetRuntimePaths sets runtime path values shown in settings.
func (m *Model) SetRuntimePaths(llamaPath, modelsPath string) {
	if llamaPath != "" {
		m.llamaPath = llamaPath
	}
	if modelsPath != "" {
		m.modelsPath = modelsPath
	}
}

// SetGenerationSettings sets generation settings shown in settings.
func (m *Model) SetGenerationSettings(temperature, topP float64, topK int, repeatPenalty float64, maxTokens int) {
	m.temperature = temperature
	m.topP = topP
	m.topK = topK
	m.repeatPenalty = repeatPenalty
	m.maxTokens = maxTokens
}

// SetFontSize sets UI font scale (terminal-friendly approximation).
func (m *Model) SetFontSize(size int) {
	if size < 1 {
		size = 1
	}
	if size > 4 {
		size = 4
	}
	m.fontSize = size
}

// SetModels sets the available models.
func (m *Model) SetModels(modelList []models.ModelInfo) {
	m.availableModels = modelList
}

// SetServerReady marks the server as ready.
func (m *Model) SetServerReady() {
	m.serverReady = true
	m.showOverlay = OverlayNone
}

// ShowWelcome shows the welcome screen.
func (m *Model) ShowWelcome() {
	m.showOverlay = OverlayWelcome
}

// ShowLoading shows the loading overlay.
func (m *Model) ShowLoading() {
	m.showOverlay = OverlayLoading
}

// SetError sets the current UI error message.
func (m *Model) SetError(err error) {
	m.err = err
}

func (m *Model) openSettings() {
	m.settingsEditing = false
	m.settingsField = 0
	m.settingsModelIdx = m.selectedModel
	m.settingsInputs = make([]textinput.Model, 8)

	newInput := func(value string) textinput.Model {
		ti := textinput.New()
		ti.SetValue(value)
		ti.CharLimit = 2048
		ti.Width = maxInt(18, m.width-30)
		return ti
	}

	m.settingsInputs[0] = newInput(m.llamaPath)
	m.settingsInputs[1] = newInput(m.modelsPath)
	m.settingsInputs[2] = newInput(fmt.Sprintf("%d", m.fontSize))
	m.settingsInputs[3] = newInput(fmt.Sprintf("%.2f", m.temperature))
	m.settingsInputs[4] = newInput(fmt.Sprintf("%.2f", m.topP))
	m.settingsInputs[5] = newInput(fmt.Sprintf("%d", m.topK))
	m.settingsInputs[6] = newInput(fmt.Sprintf("%.2f", m.repeatPenalty))
	m.settingsInputs[7] = newInput(fmt.Sprintf("%d", m.maxTokens))
	m.showOverlay = OverlaySettings
}

func (m *Model) closeSettings() {
	m.settingsEditing = false
	m.showOverlay = OverlayNone
}

// AddChatMessage adds a message to the chat view.
func (m *Model) AddChatMessage(role, content string) {
	m.messages = append(m.messages, ChatMessage{Role: role, Content: content})
}

// SetStreaming sets the streaming state.
func (m *Model) SetStreaming(streaming bool) {
	m.streaming = streaming
}

// AppendStreamContent appends content to the current streaming message.
func (m *Model) AppendStreamContent(content string) {
	m.streamBuffer += content
}

// FinishStream finishes the current stream and adds the message.
func (m *Model) FinishStream() string {
	content := m.streamBuffer
	m.streamBuffer = ""
	return content
}

// Init initializes the model (Bubble Tea interface).
func (m Model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.spinner.Tick)
}

// Ensure Model implements tea.Model at compile time.
var _ tea.Model = Model{}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
