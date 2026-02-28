package ui

import (
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
		m.textarea.Focus()
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
		m.textarea.Focus()

	case ServerReadyMsg:
		m.serverReady = true
		m.showOverlay = OverlayNone

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
			m.textarea.Focus()
			m.updateViewportContent()
		} else if m.showOverlay != OverlayNone {
			m.showOverlay = OverlayNone
		}
		return m, nil

	case key.Matches(msg, m.keys.NewLine):
		if m.streaming || m.showOverlay != OverlayNone {
			return m, nil
		}
		value := m.textarea.Value()
		m.textarea.SetValue(value + "\n")
		return m, nil

	case key.Matches(msg, m.keys.Send):
		if m.streaming || m.showOverlay != OverlayNone {
			return m, nil
		}
		text := m.textarea.Value()
		if len(text) == 0 {
			return m, nil
		}
		m.textarea.Reset()
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
		m.messages = nil
		m.updateViewportContent()
		return m, nil

	case key.Matches(msg, m.keys.ModelPicker):
		if !m.streaming {
			m.showOverlay = OverlayModelPicker
		}
		return m, nil

	case key.Matches(msg, m.keys.TemplatePicker):
		if !m.streaming {
			m.showOverlay = OverlayTemplatePicker
		}
		return m, nil

	default:
		// Handle overlay-specific keys
		if m.showOverlay == OverlayModelPicker {
			return m.handleModelPickerKey(msg)
		}
		if m.showOverlay == OverlayWelcome {
			if key.Matches(msg, m.keys.Rescan) {
				// Trigger rescan
				return m, nil
			}
		}

		// Pass to textarea
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		return m, cmd
	}
}

func (m *Model) handleModelPickerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selectedModel > 0 {
			m.selectedModel--
		}
	case "down", "j":
		if m.selectedModel < len(m.availableModels)-1 {
			m.selectedModel++
		}
	case "enter":
		if m.onModelSwitch != nil {
			m.onModelSwitch(m.selectedModel)
		}
		m.showOverlay = OverlayLoading
	case "esc":
		m.showOverlay = OverlayNone
	}
	return m, nil
}
