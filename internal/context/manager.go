package context

import (
	"github.com/VectorBlue-06/open-llama/internal/templates"
)

// Manager manages the conversation context window.
type Manager struct {
	systemPrompt       string
	messages           []templates.Message
	maxTokens          int
	reserveForResponse int
}

// NewManager creates a new context manager.
func NewManager(maxTokens int, reserveForResponse int) *Manager {
	if reserveForResponse <= 0 {
		reserveForResponse = 512
	}
	return &Manager{
		maxTokens:          maxTokens,
		reserveForResponse: reserveForResponse,
	}
}

// SetSystemPrompt sets the system prompt.
func (m *Manager) SetSystemPrompt(prompt string) {
	m.systemPrompt = prompt
}

// Add appends a message to the conversation history.
func (m *Manager) Add(msg templates.Message) {
	m.messages = append(m.messages, msg)
}

// Messages returns all messages in the conversation.
func (m *Manager) Messages() []templates.Message {
	return m.messages
}

// Clear removes all messages (keeps system prompt).
func (m *Manager) Clear() {
	m.messages = nil
}

// Build constructs the full prompt string using the given template engine.
// It applies sliding-window trimming, including messages from newest to oldest.
func (m *Manager) Build(engine *templates.Engine) string {
	available := m.maxTokens - m.reserveForResponse

	// System prompt always included
	prompt := engine.FormatSystem(m.systemPrompt)
	used := EstimateTokens(prompt)

	// Walk messages from newest to oldest, include as many as fit
	var included []templates.Message
	for i := len(m.messages) - 1; i >= 0; i-- {
		formatted := engine.FormatMessage(m.messages[i])
		cost := EstimateTokens(formatted)
		if used+cost > available {
			break
		}
		included = append([]templates.Message{m.messages[i]}, included...)
		used += cost
	}

	// Build final prompt string
	for _, msg := range included {
		prompt += engine.FormatMessage(msg)
	}
	prompt += engine.AssistantPrefix()

	return prompt
}

// Stats returns current context usage statistics.
func (m *Manager) Stats(engine *templates.Engine) ContextStats {
	prompt := m.Build(engine)
	tokensUsed := EstimateTokens(prompt)
	messagesIncluded := m.countIncluded(engine)

	pct := 0.0
	if m.maxTokens > 0 {
		pct = float64(tokensUsed) / float64(m.maxTokens) * 100
	}

	return ContextStats{
		TokensUsed:       tokensUsed,
		TokensMax:        m.maxTokens,
		MessagesTotal:    len(m.messages),
		MessagesIncluded: messagesIncluded,
		PercentUsed:      pct,
	}
}

func (m *Manager) countIncluded(engine *templates.Engine) int {
	available := m.maxTokens - m.reserveForResponse
	used := EstimateTokens(engine.FormatSystem(m.systemPrompt))
	count := 0

	for i := len(m.messages) - 1; i >= 0; i-- {
		formatted := engine.FormatMessage(m.messages[i])
		cost := EstimateTokens(formatted)
		if used+cost > available {
			break
		}
		used += cost
		count++
	}
	return count
}

// ContextStats holds context usage information.
type ContextStats struct {
	TokensUsed       int
	TokensMax        int
	MessagesTotal    int
	MessagesIncluded int
	PercentUsed      float64
}

// SetMaxTokens updates the maximum token count.
func (m *Manager) SetMaxTokens(max int) {
	m.maxTokens = max
}
