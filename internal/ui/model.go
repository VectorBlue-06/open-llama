package ui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
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
)

// ChatMessage represents a displayed chat message.
type ChatMessage struct {
	Role    string
	Content string
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
	showOverlay   OverlayType
	err           error
	messages      []ChatMessage
	streamBuffer  strings.Builder
	lastRender    time.Time
	ready         bool
	serverReady   bool

	// Models
	availableModels []models.ModelInfo
	selectedModel   int

	// Callbacks
	onSend        func(string)
	onQuit        func()
	onNewChat     func()
	onModelSwitch func(int)
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
		keys:       DefaultKeyMap(),
		textarea:   ta,
		spinner:    sp,
		metrics:    metricsCollector,
		messages:   []ChatMessage{},
		lastRender: time.Now(),
	}
}

// SetCallbacks sets the callback functions.
func (m *Model) SetCallbacks(onSend func(string), onQuit func(), onNewChat func(), onModelSwitch func(int)) {
	m.onSend = onSend
	m.onQuit = onQuit
	m.onNewChat = onNewChat
	m.onModelSwitch = onModelSwitch
}

// SetModelInfo sets the current model display info.
func (m *Model) SetModelInfo(name string, templateName string) {
	m.modelName = name
	m.templateName = templateName
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
	m.streamBuffer.WriteString(content)
}

// FinishStream finishes the current stream and adds the message.
func (m *Model) FinishStream() string {
	content := m.streamBuffer.String()
	m.streamBuffer.Reset()
	return content
}

// Init initializes the model (Bubble Tea interface).
func (m Model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.spinner.Tick)
}

// Ensure Model implements tea.Model at compile time.
var _ tea.Model = Model{}
