package ui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all keyboard shortcuts.
type KeyMap struct {
	Send           key.Binding
	NewLine        key.Binding
	NewChat        key.Binding
	Settings       key.Binding
	ModelPicker    key.Binding
	TemplatePicker key.Binding
	SaveSession    key.Binding
	Quit           key.Binding
	ForceQuit      key.Binding
	Cancel         key.Binding
	ScrollUp       key.Binding
	ScrollDown     key.Binding
	PageUp         key.Binding
	PageDown       key.Binding
	Home           key.Binding
	End            key.Binding
	ClearScreen    key.Binding
	Tab            key.Binding
	Rescan         key.Binding
}

// DefaultKeyMap returns the default key bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Send: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "send"),
		),
		NewLine: key.NewBinding(
			key.WithKeys("shift+enter", "ctrl+enter", "alt+enter"),
			key.WithHelp("shift/ctrl+enter", "newline"),
		),
		NewChat: key.NewBinding(
			key.WithKeys("ctrl+n"),
			key.WithHelp("ctrl+n", "new chat"),
		),
		Settings: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "settings"),
		),
		ModelPicker: key.NewBinding(
			key.WithKeys("ctrl+m", "f2", "alt+m"),
			key.WithHelp("ctrl+m/f2", "models"),
		),
		TemplatePicker: key.NewBinding(
			key.WithKeys("ctrl+t"),
			key.WithHelp("ctrl+t", "templates"),
		),
		SaveSession: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+q"),
			key.WithHelp("ctrl+q", "quit"),
		),
		ForceQuit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		ScrollUp: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", "scroll up"),
		),
		ScrollDown: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("↓", "scroll down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup"),
			key.WithHelp("pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown"),
			key.WithHelp("pgdn", "page down"),
		),
		Home: key.NewBinding(
			key.WithKeys("home"),
			key.WithHelp("home", "top"),
		),
		End: key.NewBinding(
			key.WithKeys("end"),
			key.WithHelp("end", "bottom"),
		),
		ClearScreen: key.NewBinding(
			key.WithKeys("ctrl+l"),
			key.WithHelp("ctrl+l", "clear"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "mode"),
		),
		Rescan: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "rescan"),
		),
	}
}
