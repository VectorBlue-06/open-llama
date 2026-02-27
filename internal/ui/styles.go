package ui

import "github.com/charmbracelet/lipgloss"

var (
	ColorPrimary   = lipgloss.Color("#7C3AED")
	ColorSecondary = lipgloss.Color("#6B7280")
	ColorUser      = lipgloss.Color("#3B82F6")
	ColorAssistant = lipgloss.Color("#10B981")
	ColorError     = lipgloss.Color("#EF4444")
	ColorDim       = lipgloss.Color("#4B5563")
	ColorHighlight = lipgloss.Color("#F59E0B")

	TopBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#1F2937")).
			Foreground(lipgloss.Color("#F9FAFB")).
			Padding(0, 1).
			Bold(true)

	StatusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#1F2937")).
			Foreground(ColorDim).
			Padding(0, 1)

	UserMsgStyle = lipgloss.NewStyle().
			Foreground(ColorUser).
			Bold(true)

	AssistantMsgStyle = lipgloss.NewStyle().
				Foreground(ColorAssistant)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)

	InputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(0, 1)

	ChatViewStyle = lipgloss.NewStyle().
			Padding(1, 2)

	WelcomeStyle = lipgloss.NewStyle().
			Align(lipgloss.Center).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(2, 4)

	SpinnerStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary)

	DimStyle = lipgloss.NewStyle().
			Foreground(ColorDim)
)
