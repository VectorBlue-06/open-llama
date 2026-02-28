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

	switch m.showOverlay {
	case OverlayWelcome:
		return m.viewWelcome()
	case OverlayLoading:
		return m.viewLoading()
	case OverlayModelPicker:
		return m.viewWithOverlay(m.viewModelPicker())
	case OverlayTemplatePicker:
		return m.viewWithOverlay(m.viewTemplatePicker())
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
	hints := "Ctrl+N New │ Ctrl+M Model │ Ctrl+T Template │ Ctrl+Q Quit"
	if m.streaming {
		hints = "Esc Cancel │ Streaming..."
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
			sb.WriteString(UserTextStyle.Render(msg.Content))
			sb.WriteString("\n\n")
		case "assistant":
			sb.WriteString(AssistantMsgStyle.Render("Assistant:"))
			sb.WriteString("\n")
			sb.WriteString(AssistantTextStyle.Render(msg.Content))
			sb.WriteString("\n\n")
		}
	}

	// Show streaming content
	if m.streaming && m.streamBuffer != "" {
		sb.WriteString(AssistantMsgStyle.Render("Assistant:"))
		sb.WriteString("\n")
		sb.WriteString(AssistantTextStyle.Render(m.streamBuffer))
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

func (m Model) viewWithOverlay(overlay string) string {
	// Render chat behind overlay
	overlayRendered := WelcomeStyle.Render(overlay)

	// Center overlay over background
	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		overlayRendered,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("#000000")),
	)
}
