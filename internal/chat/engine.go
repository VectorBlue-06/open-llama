package chat

import (
	"context"
	"net/http"
	"time"

	"github.com/VectorBlue-06/open-llama/internal/templates"
)

// EngineConfig holds configuration for the chat engine.
type EngineConfig struct {
	Temperature   float64
	TopP          float64
	TopK          int
	RepeatPenalty float64
	MaxTokens     int
}

// Engine manages conversation state and communicates with the server.
type Engine struct {
	serverPort int
	config     EngineConfig
	template   *templates.Engine
	messages   []Message
	client     *http.Client
	cancelFunc context.CancelFunc
}

// NewEngine creates a new chat engine.
func NewEngine(serverPort int, cfg EngineConfig, tmpl *templates.Engine) *Engine {
	return &Engine{
		serverPort: serverPort,
		config:     cfg,
		template:   tmpl,
		client: &http.Client{
			Timeout: 0, // No timeout for streaming
			Transport: &http.Transport{
				MaxIdleConns:        1,
				MaxIdleConnsPerHost: 1,
				IdleConnTimeout:     120 * time.Second,
				DisableKeepAlives:   false,
			},
		},
	}
}

// Send sends a user message and returns a channel of streamed tokens.
func (e *Engine) Send(ctx context.Context, prompt string) (<-chan StreamToken, error) {
	// Create a cancellable context
	ctx, cancel := context.WithCancel(ctx)
	e.cancelFunc = cancel

	req := CompletionRequest{
		Prompt:        prompt,
		Stream:        true,
		Temperature:   e.config.Temperature,
		TopP:          e.config.TopP,
		TopK:          e.config.TopK,
		RepeatPenalty: e.config.RepeatPenalty,
		NPredict:      e.config.MaxTokens,
		Stop:          e.template.StopTokens(),
	}

	return streamCompletion(ctx, e.client, e.serverPort, req)
}

// Cancel cancels the current streaming request.
func (e *Engine) Cancel() {
	if e.cancelFunc != nil {
		e.cancelFunc()
		e.cancelFunc = nil
	}
}

// AddMessage adds a message to the engine's history.
func (e *Engine) AddMessage(msg Message) {
	e.messages = append(e.messages, msg)
}

// Messages returns all messages.
func (e *Engine) Messages() []Message {
	return e.messages
}

// Reset clears the conversation history.
func (e *Engine) Reset() {
	e.messages = nil
}

// SetTemplate changes the template engine.
func (e *Engine) SetTemplate(t *templates.Engine) {
	e.template = t
}

// SetServerPort updates the server port (for model switching).
func (e *Engine) SetServerPort(port int) {
	e.serverPort = port
}
