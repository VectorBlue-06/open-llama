package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/VectorBlue-06/open-llama/internal/models"
)

// View renders the UI (Bubble Tea interface).
func (m Model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	if m.mode == ModeStartup && m.showOverlay == OverlayNone {
		return m.viewStartup()
	}

	switch m.showOverlay {
	case OverlayWelcome:
		return m.viewWelcome()
	case OverlayLoading:
		return m.viewLoading()
	case OverlayModelPicker:
		return m.viewWithOverlay(m.viewModelPicker())
	case OverlayTemplatePicker:
		return m.viewWithOverlay(m.viewTemplatePicker())
	case OverlaySettings:
		if m.mode == ModeStartup {
			return m.viewWithOverlay(m.viewSettingsOverlay(), m.viewStartup())
		}
		return m.viewWithOverlay(m.viewSettingsOverlay(), m.viewChat())
	}

	return m.viewChat()
}

func (m Model) viewChat() string {
	topBar := m.renderTopBar()
	chatView := m.viewport.View()
	input := InputStyle.Width(m.width - 4).Render(m.textarea.View())
	statusBar := m.renderStatusBar()

	return lipgloss.JoinVertical(lipgloss.Left,
		topBar,
		chatView,
		input,
		statusBar,
	)
}

func (m Model) renderTopBar() string {
	modelStr := m.modelName
	if modelStr == "" {
		modelStr = "No model"
	}
	if runes := []rune(modelStr); len(runes) > 20 {
		modelStr = string(runes[:20])
	}

	tmplStr := m.templateName
	if tmplStr == "" {
		tmplStr = "—"
	}

	speedStr := "— t/s"
	ctxStr := "CTX —"
	gpuStr := ""

	if m.metrics != nil {
		speedStr = m.metrics.SpeedString()
		ctxStr = m.metrics.ContextString()
		snap := m.metrics.Snapshot()
		if snap.GPUActive {
			gpuStr = " │ GPU"
		}
	}

	content := fmt.Sprintf(" %s │ %s │ %s │ %s%s",
		modelStr, tmplStr, ctxStr, speedStr, gpuStr)

	return TopBarStyle.Width(m.width).Render(content)
}

func (m Model) renderStatusBar() string {
	hints := "Esc Menu │ Tab Settings │ Shift+Enter New Line │ Ctrl+Q Quit"
	if m.streaming {
		hints = "Esc Cancel │ Streaming..."
	}
	if m.mode == ModeStartup {
		hints = "Enter Start Chat │ Tab Settings │ Esc Menu │ Ctrl+Q Quit"
	}
	if m.err != nil {
		hints = ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}
	return StatusBarStyle.Width(m.width).Render(hints)
}

func (m *Model) updateViewportContent() {
	var sb strings.Builder

	for _, msg := range m.messages {
		switch msg.Role {
		case "user":
			sb.WriteString(UserMsgStyle.Render("You:"))
			sb.WriteString("\n")
			sb.WriteString(m.scaleText(UserTextStyle, msg.Content))
			sb.WriteString("\n\n")
		case "assistant":
			sb.WriteString(AssistantMsgStyle.Render("Assistant:"))
			sb.WriteString("\n")
			sb.WriteString(m.scaleText(AssistantTextStyle, msg.Content))
			sb.WriteString("\n\n")
		}
	}

	// Show streaming content
	if m.streaming && m.streamBuffer != "" {
		sb.WriteString(AssistantMsgStyle.Render("Assistant:"))
		sb.WriteString("\n")
		sb.WriteString(m.scaleText(AssistantTextStyle, m.streamBuffer))
		sb.WriteString("█\n")
	}

	m.viewport.SetContent(sb.String())
	m.viewport.GotoBottom()
}

func (m Model) viewWelcome() string {
	content := lipgloss.JoinVertical(lipgloss.Center,
		"",
		lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("Welcome to OpenLlama!"),
		"",
		"No models found.",
		"",
		"Place .gguf model files in:",
		DimStyle.Render("runtime/models/ (next to openllama)"),
		"",
		"Recommended starter models:",
		DimStyle.Render("• Mistral 7B Q4_K_M  (~4.4 GB RAM)"),
		DimStyle.Render("• Llama 3 8B Q4_K_M  (~5.0 GB RAM)"),
		DimStyle.Render("• Phi-3 Mini Q4_K_M  (~2.4 GB RAM)"),
		"",
		DimStyle.Render("Download from: https://huggingface.co"),
		"",
		DimStyle.Render("Press 'r' to rescan  |  'q' to quit"),
	)

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		WelcomeStyle.Render(content),
	)
}

func (m Model) viewLoading() string {
	content := fmt.Sprintf("\n  %s Loading model...\n  This may take a moment.\n", m.spinner.View())
	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		WelcomeStyle.Render(content),
	)
}

func (m Model) viewStartup() string {
	ramInfo := m.ramInfo
	if ramInfo == "" {
		ramInfo = "RAM available: —"
	}
	gpuInfo := m.gpuInfo
	if gpuInfo == "" {
		gpuInfo = "GPU: —"
	}
	selectedModel := m.modelName
	if selectedModel == "" {
		selectedModel = "No model selected"
	}

	searchBox := InputStyle.Width(maxInt(30, m.width/2)).Render(m.textarea.View())

	content := lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("Welcome to OpenLlama"),
		"",
		DimStyle.Render(ramInfo),
		DimStyle.Render(gpuInfo),
		DimStyle.Render("Selected model: "+selectedModel),
		"",
		searchBox,
		"",
		DimStyle.Render("Type a prompt and press Enter to start chat"),
	)

	statusBar := m.renderStatusBar()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.Place(m.width, m.height-1, lipgloss.Center, lipgloss.Center, WelcomeStyle.Render(content)),
		statusBar,
	)
}

func (m Model) viewModelPicker() string {
	var sb strings.Builder
	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("Select Model"))
	sb.WriteString("\n\n")

	for i, model := range m.availableModels {
		cursor := "  "
		style := lipgloss.NewStyle()
		if i == m.selectedModel {
			cursor = "▸ "
			style = style.Bold(true).Foreground(ColorPrimary)
		}

		line := fmt.Sprintf("%s%s  %s  %s",
			cursor,
			model.Filename,
			DimStyle.Render(model.ParameterCount),
			DimStyle.Render(models.FormatSize(model.FileSize)),
		)
		sb.WriteString(style.Render(line))
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(DimStyle.Render("↑/↓ select  │  Enter confirm  │  Esc cancel"))

	return sb.String()
}

func (m Model) viewTemplatePicker() string {
	templates := []string{"ChatML", "Llama 2", "Llama 3", "Alpaca", "Minimal"}

	var sb strings.Builder
	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("Select Template"))
	sb.WriteString("\n\n")

	for _, t := range templates {
		marker := "  "
		if t == m.templateName {
			marker = "● "
		}
		sb.WriteString(fmt.Sprintf("%s%s\n", marker, t))
	}

	sb.WriteString("\n")
	sb.WriteString(DimStyle.Render("↑/↓ select  │  Enter confirm  │  Esc cancel"))

	return sb.String()
}

func (m Model) viewSettingsOverlay() string {
	var sb strings.Builder
	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("Settings"))
	sb.WriteString("\n\n")

	labels := []string{
		"llama.cpp path",
		"models path",
		"font size",
		"temperature",
		"top_p",
		"top_k",
		"repeat penalty",
		"max tokens",
	}

	for i, label := range labels {
		prefix := "  "
		if m.settingsField == i {
			prefix = "▸ "
		}
		sb.WriteString(prefix + label + ": " + m.settingsInputs[i].View() + "\n")
	}

	modelLine := "(none)"
	if len(m.availableModels) > 0 && m.settingsModelIdx >= 0 && m.settingsModelIdx < len(m.availableModels) {
		modelLine = m.availableModels[m.settingsModelIdx].Filename
	}
	prefix := "  "
	if m.settingsField == 8 {
		prefix = "▸ "
	}
	sb.WriteString(prefix + "selected model: " + modelLine + "\n")

	prefix = "  "
	if m.settingsField == 9 {
		prefix = "▸ "
	}
	sb.WriteString(prefix + "Apply and close\n\n")

	sb.WriteString(DimStyle.Render("↑/↓ move  │  Enter edit/apply  │  ←/→ model  │  Esc close"))
	return sb.String()
}

func (m Model) viewWithOverlay(overlay string, background ...string) string {
	bg := m.viewChat()
	if len(background) > 0 {
		bg = background[0]
	}

	overlayRendered := WelcomeStyle.Render(overlay)

	return bg + "\n" + lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		overlayRendered)
}

func (m Model) scaleText(style lipgloss.Style, text string) string {
	if m.fontSize <= 1 {
		return style.Render(text)
	}
	rendered := style.Bold(true).Render(text)
	if m.fontSize >= 3 {
		return rendered + "\n"
	}
	return rendered
}
