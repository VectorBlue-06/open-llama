package templates

import (
	"strings"
	"testing"
)

func TestFormatSystemChatML(t *testing.T) {
	e := NewEngine(ChatML)
	result := e.FormatSystem("You are helpful.")
	expected := "<|im_start|>system\nYou are helpful.<|im_end|>\n"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatSystemEmpty(t *testing.T) {
	e := NewEngine(ChatML)
	result := e.FormatSystem("")
	if result != "" {
		t.Errorf("expected empty string for empty system prompt, got %q", result)
	}
}

func TestFormatMessageUser(t *testing.T) {
	e := NewEngine(ChatML)
	msg := Message{Role: RoleUser, Content: "Hello!"}
	result := e.FormatMessage(msg)
	expected := "<|im_start|>user\nHello!<|im_end|>\n"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFormatMessageAssistant(t *testing.T) {
	e := NewEngine(ChatML)
	msg := Message{Role: RoleAssistant, Content: "Hi there!"}
	result := e.FormatMessage(msg)
	expected := "<|im_start|>assistant\nHi there!<|im_end|>\n"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestLlama2Template(t *testing.T) {
	e := NewEngine(Llama2)
	result := e.FormatSystem("Be helpful.")
	if !strings.Contains(result, "<<SYS>>") {
		t.Error("expected Llama 2 system markers")
	}
	if !strings.Contains(result, "<</SYS>>") {
		t.Error("expected Llama 2 system closing marker")
	}
}

func TestLlama3Template(t *testing.T) {
	e := NewEngine(Llama3)
	result := e.FormatSystem("Be helpful.")
	if !strings.Contains(result, "<|start_header_id|>system<|end_header_id|>") {
		t.Error("expected Llama 3 system header")
	}
}

func TestGetByName(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"chatml", "ChatML"},
		{"ChatML", "ChatML"},
		{"llama2", "Llama 2"},
		{"llama3", "Llama 3"},
		{"alpaca", "Alpaca"},
		{"minimal", "Minimal"},
		{"unknown", "ChatML"}, // default fallback
	}
	for _, tt := range tests {
		got := GetByName(tt.name)
		if got.Name != tt.expected {
			t.Errorf("GetByName(%q) = %q, want %q", tt.name, got.Name, tt.expected)
		}
	}
}

func TestDetectFromFilename(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"mistral-7b-chatml-q4.gguf", "ChatML"},
		{"llama-2-13b-chat.gguf", "Llama 2"},
		{"llama2-7b.gguf", "Llama 2"},
		{"llama-3-8b.gguf", "Llama 3"},
		{"llama3-instruct.gguf", "Llama 3"},
		{"alpaca-7b.gguf", "Alpaca"},
		{"mistral-7b-q4.gguf", "ChatML"}, // default
	}
	for _, tt := range tests {
		got := DetectFromFilename(tt.filename)
		if got.Name != tt.expected {
			t.Errorf("DetectFromFilename(%q) = %q, want %q", tt.filename, got.Name, tt.expected)
		}
	}
}

func TestAllBuiltinTemplatesHaveStopTokens(t *testing.T) {
	for name, tmpl := range BuiltinTemplates {
		if len(tmpl.StopTokens) == 0 {
			t.Errorf("template %q has no stop tokens", name)
		}
	}
}
