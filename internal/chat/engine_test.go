package chat

import (
	"testing"
	"time"

	"github.com/VectorBlue-06/open-llama/internal/templates"
)

func TestNewEngine(t *testing.T) {
	tmpl := templates.NewEngine(templates.ChatML)
	cfg := EngineConfig{
		Temperature:   0.7,
		TopP:          0.9,
		TopK:          40,
		RepeatPenalty: 1.1,
		MaxTokens:     2048,
	}

	e := NewEngine(8080, cfg, tmpl)
	if e == nil {
		t.Fatal("expected non-nil engine")
	}
}

func TestEngineAddMessage(t *testing.T) {
	tmpl := templates.NewEngine(templates.ChatML)
	cfg := EngineConfig{}
	e := NewEngine(8080, cfg, tmpl)

	e.AddMessage(Message{
		Role:      RoleUser,
		Content:   "Hello",
		Timestamp: time.Now(),
	})

	msgs := e.Messages()
	if len(msgs) != 1 {
		t.Errorf("expected 1 message, got %d", len(msgs))
	}
	if msgs[0].Content != "Hello" {
		t.Errorf("expected content 'Hello', got %q", msgs[0].Content)
	}
}

func TestEngineReset(t *testing.T) {
	tmpl := templates.NewEngine(templates.ChatML)
	cfg := EngineConfig{}
	e := NewEngine(8080, cfg, tmpl)

	e.AddMessage(Message{Role: RoleUser, Content: "Hello"})
	e.Reset()

	if len(e.Messages()) != 0 {
		t.Error("expected 0 messages after reset")
	}
}

func TestEngineCancel(t *testing.T) {
	tmpl := templates.NewEngine(templates.ChatML)
	cfg := EngineConfig{}
	e := NewEngine(8080, cfg, tmpl)

	// Cancel should not panic even without an active request
	e.Cancel()
}
