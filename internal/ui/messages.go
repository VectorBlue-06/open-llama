package ui

import (
	"github.com/VectorBlue-06/open-llama/internal/chat"
	"github.com/VectorBlue-06/open-llama/internal/models"
)

// StreamChunkMsg is sent when a stream chunk arrives from the server.
type StreamChunkMsg struct {
	Content string
}

// StreamDoneMsg is sent when streaming is complete.
type StreamDoneMsg struct {
	FullContent string
	Timings     *chat.Timings
}

// StreamErrorMsg is sent when an error occurs during streaming.
type StreamErrorMsg struct {
	Err error
}

// ServerReadyMsg is sent when server health check completes.
type ServerReadyMsg struct{}

// ServerFailedMsg is sent when server fails to start.
type ServerFailedMsg struct {
	Err error
}

// ModelsScanCompleteMsg is sent when model scan completes.
type ModelsScanCompleteMsg struct {
	Models []models.ModelInfo
}

// TickMsg is sent on a timer to throttle UI redraws.
type TickMsg struct{}
