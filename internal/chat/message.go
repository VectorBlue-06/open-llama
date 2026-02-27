package chat

import "time"

// Role represents a message role.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// Message represents a single chat message.
type Message struct {
	Role      Role      `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// StreamToken represents a token received during streaming.
type StreamToken struct {
	Content string   `json:"content"`
	Stop    bool     `json:"stop"`
	Timings *Timings `json:"timings,omitempty"`
}

// Timings holds performance timing information from the server.
type Timings struct {
	PredictedPerSecond float64 `json:"predicted_per_second"`
	PromptTokens       int     `json:"prompt_n"`
	PredictedTokens    int     `json:"predicted_n"`
	PromptMS           float64 `json:"prompt_ms"`
	PredictedMS        float64 `json:"predicted_ms"`
}
