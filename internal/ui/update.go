package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// Update handles messages (Bubble Tea interface).
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case tea.WindowSizeMsg:
		m.width = maxInt(20, msg.Width)
		m.height = maxInt(10, msg.Height)

		viewportWidth := maxInt(1, m.width)
		viewportHeight := maxInt(1, m.height-8)
		textareaWidth := maxInt(10, m.width-4)

		if !m.ready {
			m.viewport = viewport.New(viewportWidth, viewportHeight)
			m.viewport.HighPerformanceRendering = false
			m.ready = true
		} else {
			m.viewport.Width = viewportWidth
			m.viewport.Height = viewportHeight
		}
		m.textarea.SetWidth(textareaWidth)
		m.adjustTextareaHeight()
		for i := range m.settingsInputs {
			m.settingsInputs[i].Width = maxInt(18, m.width-30)
		}
		m.updateViewportContent()
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case StreamChunkMsg:
		m.AppendStreamContent(msg.Content)
		m.updateViewportContent()

	case StreamDoneMsg:
		content := m.FinishStream()
		if content == "" {
			content = msg.FullContent
		}
		m.AddChatMessage("assistant", content)
		m.streaming = false
		m.setInputMode(InputModeInsert)
		m.adjustTextareaHeight()
		m.updateViewportContent()

		if msg.Timings != nil && m.metrics != nil {
			m.metrics.UpdateFromResponse(
				msg.Timings.PredictedPerSecond,
				msg.Timings.PromptTokens,
				msg.Timings.PredictedTokens,
			)
		}

	case StreamErrorMsg:
		m.streaming = false
		m.err = msg.Err
		m.setInputMode(InputModeInsert)
		m.adjustTextareaHeight()

	case ServerReadyMsg:
		m.serverReady = true
		if m.showOverlay == OverlayLoading {
			m.showOverlay = OverlayNone
		}

	case ServerFailedMsg:
		m.err = msg.Err
		m.showOverlay = OverlayNone

	case ModelsScanCompleteMsg:
		m.availableModels = msg.Models
		if len(msg.Models) == 0 {
			m.showOverlay = OverlayWelcome
		}
	}

	// Update textarea
	if !m.streaming && m.showOverlay == OverlayNone {
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.showOverlay == OverlaySettings || m.showOverlay == OverlaySettingsConfirm {
		return m.handleSettingsKey(msg)
	}
	if m.showOverlay == OverlayModelPicker {
		return m.handleModelPickerKey(msg)
	}

	switch {
	case key.Matches(msg, m.keys.Quit), key.Matches(msg, m.keys.ForceQuit):
		if m.onQuit != nil {
			m.onQuit()
		}
		return m, tea.Quit

	case key.Matches(msg, m.keys.Cancel):
		if m.streaming {
			m.streaming = false
			content := m.FinishStream()
			if content != "" {
				m.AddChatMessage("assistant", content+" [interrupted]")
			}
			m.setInputMode(InputModeInsert)
			m.updateViewportContent()
		} else if m.showOverlay != OverlayNone {
			m.showOverlay = OverlayNone
		} else {
			m.openSettings()
		}
		return m, nil

	case key.Matches(msg, m.keys.ModelPicker), key.Matches(msg, m.keys.Tab):
		if key.Matches(msg, m.keys.Tab) {
			if m.inputMode == InputModeInsert {
				m.setInputMode(InputModeNormal)
			} else {
				m.setInputMode(InputModeInsert)
			}
			return m, nil
		}
		if !m.streaming {
			if len(m.availableModels) == 0 {
				m.err = fmt.Errorf("no models available")
				return m, nil
			}
			m.openModelPickerAfterSettings = false
			m.showOverlay = OverlayModelPicker
		}
		return m, nil

	case key.Matches(msg, m.keys.Settings):
		if !m.streaming {
			m.openSettings()
		}
		return m, nil

	case key.Matches(msg, m.keys.NewLine):
		if m.streaming || m.showOverlay != OverlayNone || m.inputMode != InputModeInsert {
			return m, nil
		}
		value := m.textarea.Value()
		m.textarea.SetValue(value + "\n")
		m.adjustTextareaHeight()
		return m, nil

	case key.Matches(msg, m.keys.Send):
		if m.streaming || m.showOverlay != OverlayNone || m.inputMode != InputModeInsert {
			return m, nil
		}
		text := strings.TrimSpace(m.textarea.Value())
		if len(text) == 0 {
			return m, nil
		}
		if m.mode == ModeStartup {
			m.mode = ModeChat
		}
		m.textarea.Reset()
		m.adjustTextareaHeight()
		m.AddChatMessage("user", text)
		m.streaming = true
		m.streamBuffer = ""
		m.updateViewportContent()

		if m.onSend != nil {
			m.onSend(text)
		}
		return m, nil

	case key.Matches(msg, m.keys.NewChat):
		if m.onNewChat != nil {
			m.onNewChat()
		}
		m.mode = ModeChat
		m.messages = nil
		m.updateViewportContent()
		return m, nil

	case key.Matches(msg, m.keys.TemplatePicker):
		if !m.streaming {
			m.showOverlay = OverlayTemplatePicker
		}
		return m, nil

	default:
		if m.showOverlay == OverlayWelcome {
			if key.Matches(msg, m.keys.Rescan) {
				// Trigger rescan
				return m, nil
			}
		}

		if m.showOverlay == OverlayNone && m.inputMode == InputModeNormal {
			switch msg.String() {
			case "j":
				m.viewport.LineUp(1)
				return m, nil
			case "k":
				m.viewport.LineDown(1)
				return m, nil
			case "up":
				m.viewport.LineUp(1)
				return m, nil
			case "down":
				m.viewport.LineDown(1)
				return m, nil
			case "pgup":
				m.viewport.HalfViewUp()
				return m, nil
			case "pgdown":
				m.viewport.HalfViewDown()
				return m, nil
			}
			return m, nil
		}

		// Pass to textarea
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		m.adjustTextareaHeight()
		return m, cmd
	}
}

func (m *Model) handleSettingsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.showOverlay == OverlaySettingsConfirm {
		switch msg.String() {
		case "left", "h", "right", "l":
			if m.settingsConfirmChoice == 0 {
				m.settingsConfirmChoice = 1
			} else {
				m.settingsConfirmChoice = 0
			}
			return m, nil
		case "esc":
			m.showOverlay = OverlaySettings
			return m, nil
		case "enter":
			if m.settingsConfirmChoice == 1 {
				if err := m.applySettings(); err != nil {
					m.err = err
					m.showOverlay = OverlaySettings
					return m, nil
				}
			}
			openModelPicker := m.openModelPickerAfterSettings
			m.openModelPickerAfterSettings = false
			m.closeSettings()
			if openModelPicker {
				m.showOverlay = OverlayModelPicker
			}
			return m, nil
		}
		return m, nil
	}

	const modelFieldIndex = 7
	const applyFieldIndex = 8

	if key.Matches(msg, m.keys.Cancel) {
		if m.settingsEditing {
			m.settingsEditing = false
			return m, nil
		}
		if m.settingsDirty() {
			m.settingsConfirmChoice = 0
			m.showOverlay = OverlaySettingsConfirm
			return m, nil
		}
		m.closeSettings()
		return m, nil
	}

	if key.Matches(msg, m.keys.Settings) || key.Matches(msg, m.keys.ModelPicker) || key.Matches(msg, m.keys.Tab) {
		if key.Matches(msg, m.keys.ModelPicker) {
			m.openModelPickerAfterSettings = true
			if m.settingsDirty() {
				m.settingsConfirmChoice = 0
				m.showOverlay = OverlaySettingsConfirm
				return m, nil
			}
			m.closeSettings()
			m.showOverlay = OverlayModelPicker
			return m, nil
		}
		if key.Matches(msg, m.keys.Tab) {
			if m.settingsDirty() {
				m.settingsConfirmChoice = 0
				m.showOverlay = OverlaySettingsConfirm
				return m, nil
			}
			m.openModelPickerAfterSettings = false
			m.closeSettings()
			if m.inputMode == InputModeInsert {
				m.setInputMode(InputModeNormal)
			} else {
				m.setInputMode(InputModeInsert)
			}
			return m, nil
		}
		if m.settingsDirty() {
			m.settingsConfirmChoice = 0
			m.showOverlay = OverlaySettingsConfirm
			return m, nil
		}
		m.closeSettings()
		m.openModelPickerAfterSettings = false
		return m, nil
	}

	if m.settingsEditing {
		if m.settingsField >= 0 && m.settingsField < len(m.settingsInputs) {
			var cmd tea.Cmd
			m.settingsInputs[m.settingsField], cmd = m.settingsInputs[m.settingsField].Update(msg)
			if msg.String() == "enter" {
				m.settingsEditing = false
				return m, nil
			}
			return m, cmd
		}
		m.settingsEditing = false
	}

	switch msg.String() {
	case "up", "k":
		if m.settingsField > 0 {
			m.settingsField--
		}
		return m, nil
	case "down", "j":
		if m.settingsField < applyFieldIndex {
			m.settingsField++
		}
		return m, nil
	case "left", "h":
		if m.settingsField == modelFieldIndex && len(m.availableModels) > 0 && m.settingsModelIdx > 0 {
			m.settingsModelIdx--
		}
		return m, nil
	case "right", "l":
		if m.settingsField == modelFieldIndex && len(m.availableModels) > 0 && m.settingsModelIdx < len(m.availableModels)-1 {
			m.settingsModelIdx++
		}
		return m, nil
	case "enter":
		if m.settingsField >= 0 && m.settingsField < len(m.settingsInputs) {
			m.settingsEditing = true
			m.settingsInputs[m.settingsField].Focus()
			return m, nil
		}
		if m.settingsField == applyFieldIndex {
			m.settingsConfirmChoice = 1
			m.showOverlay = OverlaySettingsConfirm
			return m, nil
		}
	}

	return m, nil
}

func (m *Model) applySettings() error {
	parseInt := func(v string, field string) (int, error) {
		n, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return 0, fmt.Errorf("invalid %s", field)
		}
		return n, nil
	}
	parseFloat := func(v string, field string) (float64, error) {
		n, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		if err != nil {
			return 0, fmt.Errorf("invalid %s", field)
		}
		return n, nil
	}

	temp, err := parseFloat(m.settingsInputs[2].Value(), "temperature")
	if err != nil {
		return err
	}
	topP, err := parseFloat(m.settingsInputs[3].Value(), "top_p")
	if err != nil {
		return err
	}
	topK, err := parseInt(m.settingsInputs[4].Value(), "top_k")
	if err != nil {
		return err
	}
	repeatPenalty, err := parseFloat(m.settingsInputs[5].Value(), "repeat penalty")
	if err != nil {
		return err
	}
	maxTokens, err := parseInt(m.settingsInputs[6].Value(), "max tokens")
	if err != nil {
		return err
	}

	m.llamaPath = strings.TrimSpace(m.settingsInputs[0].Value())
	m.modelsPath = strings.TrimSpace(m.settingsInputs[1].Value())
	m.temperature = temp
	m.topP = topP
	m.topK = topK
	m.repeatPenalty = repeatPenalty
	m.maxTokens = maxTokens
	if len(m.availableModels) > 0 && m.settingsModelIdx >= 0 && m.settingsModelIdx < len(m.availableModels) {
		m.selectedModel = m.settingsModelIdx
	}

	if m.onApplySettings != nil {
		m.onApplySettings(SettingsUpdate{
			LlamaPath:     m.llamaPath,
			ModelsPath:    m.modelsPath,
			Temperature:   m.temperature,
			TopP:          m.topP,
			TopK:          m.topK,
			RepeatPenalty: m.repeatPenalty,
			MaxTokens:     m.maxTokens,
			SelectedModel: m.selectedModel,
		})
	}

	return nil
}

func (m *Model) handleModelPickerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if len(m.availableModels) == 0 {
		m.err = fmt.Errorf("no models available")
		m.showOverlay = OverlayNone
		return m, nil
	}

	switch msg.String() {
	case "up", "k":
		if m.selectedModel > 0 {
			m.selectedModel--
		}
	case "down", "j":
		if m.selectedModel < len(m.availableModels)-1 {
			m.selectedModel++
		}
	case "enter", "right", "l":
		if m.onModelSwitch != nil {
			m.onModelSwitch(m.selectedModel)
			m.showOverlay = OverlayLoading
			return m, nil
		}
		m.showOverlay = OverlayNone
	case "esc", "left", "h":
		m.showOverlay = OverlayNone
	}
	return m, nil
}

func (m *Model) settingsDirty() bool {
	if len(m.settingsOriginalValues) != len(m.settingsInputs) || len(m.settingsOriginalValues) == 0 {
		return false
	}
	for i := range m.settingsInputs {
		if strings.TrimSpace(m.settingsInputs[i].Value()) != strings.TrimSpace(m.settingsOriginalValues[i]) {
			return true
		}
	}
	return m.settingsModelIdx != m.settingsOriginalModel
}

func (m *Model) adjustTextareaHeight() {
	lines := strings.Count(m.textarea.Value(), "\n") + 1
	if lines < 1 {
		lines = 1
	}
	if lines > 5 {
		lines = 5
	}
	m.textarea.SetHeight(lines)
}
