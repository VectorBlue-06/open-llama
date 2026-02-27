package context

import (
	"testing"

	"github.com/VectorBlue-06/open-llama/internal/templates"
)

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"", 0},
		{"Hi", 1},       // 2 chars / 3.6 = 0.56 -> ceil = 1
		{"Hello world", 4}, // 11 chars / 3.6 = 3.06 -> ceil = 4 (actual: 3)
	}
	for _, tt := range tests {
		got := EstimateTokens(tt.input)
		if got != tt.expected {
			t.Errorf("EstimateTokens(%q) = %d, want %d", tt.input, got, tt.expected)
		}
	}
}

func TestManagerBuildSimple(t *testing.T) {
	m := NewManager(4096, 512)
	m.SetSystemPrompt("You are helpful.")

	engine := templates.NewEngine(templates.ChatML)

	m.Add(templates.Message{Role: templates.RoleUser, Content: "Hello"})
	m.Add(templates.Message{Role: templates.RoleAssistant, Content: "Hi there!"})
	m.Add(templates.Message{Role: templates.RoleUser, Content: "How are you?"})

	prompt := m.Build(engine)

	// Should contain system prompt and all messages
	if len(prompt) == 0 {
		t.Error("expected non-empty prompt")
	}
}

func TestManagerSlidingWindow(t *testing.T) {
	// Very small context to force trimming
	m := NewManager(100, 20)
	m.SetSystemPrompt("System.")

	engine := templates.NewEngine(templates.Minimal)

	// Add many messages to overflow context
	for i := 0; i < 20; i++ {
		m.Add(templates.Message{Role: templates.RoleUser, Content: "This is a message that takes up tokens."})
		m.Add(templates.Message{Role: templates.RoleAssistant, Content: "This is a response."})
	}

	stats := m.Stats(engine)

	// Should have trimmed some messages
	if stats.MessagesIncluded >= stats.MessagesTotal {
		t.Errorf("expected trimming: included %d of %d", stats.MessagesIncluded, stats.MessagesTotal)
	}
}

func TestManagerClear(t *testing.T) {
	m := NewManager(4096, 512)
	m.SetSystemPrompt("System.")
	m.Add(templates.Message{Role: templates.RoleUser, Content: "Hello"})

	if len(m.Messages()) != 1 {
		t.Error("expected 1 message")
	}

	m.Clear()

	if len(m.Messages()) != 0 {
		t.Error("expected 0 messages after clear")
	}
}

func TestManagerEmptyConversation(t *testing.T) {
	m := NewManager(4096, 512)
	m.SetSystemPrompt("You are helpful.")

	engine := templates.NewEngine(templates.ChatML)
	prompt := m.Build(engine)

	// Should still have system prompt and assistant prefix
	if len(prompt) == 0 {
		t.Error("expected non-empty prompt even with no messages")
	}
}
