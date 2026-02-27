package templates

import (
	"strings"
)

// Role represents a message role.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// Message is a minimal message interface for template formatting.
type Message struct {
	Role    Role
	Content string
}

// Engine applies templates to format prompts.
type Engine struct {
	template Template
}

// NewEngine creates a new template engine with the given template.
func NewEngine(t Template) *Engine {
	return &Engine{template: t}
}

// SetTemplate changes the active template.
func (e *Engine) SetTemplate(t Template) {
	e.template = t
}

// Template returns the current template.
func (e *Engine) Template() Template {
	return e.template
}

// FormatSystem formats the system prompt.
func (e *Engine) FormatSystem(systemPrompt string) string {
	if systemPrompt == "" {
		return ""
	}
	return e.template.SystemPrefix + systemPrompt + e.template.SystemSuffix
}

// FormatMessage formats a single message according to the template.
func (e *Engine) FormatMessage(msg Message) string {
	switch msg.Role {
	case RoleUser:
		return e.template.UserPrefix + msg.Content + e.template.UserSuffix
	case RoleAssistant:
		return e.template.AssistantPrefix + msg.Content + e.template.AssistantSuffix
	case RoleSystem:
		return e.template.SystemPrefix + msg.Content + e.template.SystemSuffix
	default:
		return msg.Content
	}
}

// AssistantPrefix returns the assistant prefix for the current template.
func (e *Engine) AssistantPrefix() string {
	return e.template.AssistantPrefix
}

// StopTokens returns the stop tokens for the current template.
func (e *Engine) StopTokens() []string {
	return e.template.StopTokens
}

// GetByName returns a template by name (case-insensitive), or the default.
func GetByName(name string) Template {
	name = strings.ToLower(name)
	if t, ok := BuiltinTemplates[name]; ok {
		return t
	}
	return ChatML
}

// DetectFromFilename attempts to match a template based on model filename.
func DetectFromFilename(filename string) Template {
	lower := strings.ToLower(filename)
	switch {
	case strings.Contains(lower, "chatml"):
		return ChatML
	case strings.Contains(lower, "llama-2") || strings.Contains(lower, "llama2"):
		return Llama2
	case strings.Contains(lower, "llama-3") || strings.Contains(lower, "llama3"):
		return Llama3
	case strings.Contains(lower, "alpaca"):
		return Alpaca
	default:
		return ChatML
	}
}
