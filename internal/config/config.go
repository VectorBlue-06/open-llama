package config

// Config holds all application configuration.
type Config struct {
	Version int `json:"version"`

	Model      ModelConfig      `json:"model"`
	Server     ServerConfig     `json:"server"`
	Generation GenerationConfig `json:"generation"`
	Template   TemplateConfig   `json:"template"`
	UI         UIConfig         `json:"ui"`
	Session    SessionConfig    `json:"session"`
	Debug      bool             `json:"debug"`
}

type ModelConfig struct {
	Default   string `json:"default"`
	LlamaPath string `json:"llama_path"`
	ModelsDir string `json:"models_dir"`
}

type ServerConfig struct {
	Host      string   `json:"host"`
	Port      int      `json:"port"`
	CtxSize   int      `json:"ctx_size"`
	BatchSize int      `json:"batch_size"`
	Threads   int      `json:"threads"`
	GPULayers int      `json:"gpu_layers"`
	ExtraArgs []string `json:"extra_args"`
}

type GenerationConfig struct {
	Temperature   float64 `json:"temperature"`
	TopP          float64 `json:"top_p"`
	TopK          int     `json:"top_k"`
	RepeatPenalty float64 `json:"repeat_penalty"`
	MaxTokens     int     `json:"max_tokens"`
}

type TemplateConfig struct {
	Default      string          `json:"default"`
	SystemPrompt string          `json:"system_prompt"`
	Custom       *CustomTemplate `json:"custom_template,omitempty"`
}

type CustomTemplate struct {
	Name            string   `json:"name"`
	SystemPrefix    string   `json:"system_prefix"`
	SystemSuffix    string   `json:"system_suffix"`
	UserPrefix      string   `json:"user_prefix"`
	UserSuffix      string   `json:"user_suffix"`
	AssistantPrefix string   `json:"assistant_prefix"`
	AssistantSuffix string   `json:"assistant_suffix"`
	StopTokens      []string `json:"stop_tokens"`
}

type UIConfig struct {
	Theme            string `json:"theme"`
	RenderThrottleMs int    `json:"render_throttle_ms"`
	ShowMetrics      bool   `json:"show_metrics"`
	ShowTimestamps   bool   `json:"show_timestamps"`
	FontSize         int    `json:"font_size"`
}

type SessionConfig struct {
	AutoSave    bool   `json:"auto_save"`
	SessionsDir string `json:"sessions_dir"`
	MaxSessions int    `json:"max_sessions"`
}

// Defaults returns a Config with all default values.
func Defaults() *Config {
	return &Config{
		Version: 1,
		Model: ModelConfig{
			Default:   "",
			LlamaPath: "runtime/llama.cpp",
			ModelsDir: "runtime/models",
		},
		Server: ServerConfig{
			Host:      "127.0.0.1",
			Port:      0,
			CtxSize:   4096,
			BatchSize: 512,
			Threads:   0,
			GPULayers: -1,
			ExtraArgs: []string{},
		},
		Generation: GenerationConfig{
			Temperature:   0.7,
			TopP:          0.9,
			TopK:          40,
			RepeatPenalty: 1.1,
			MaxTokens:     2048,
		},
		Template: TemplateConfig{
			Default:      "chatml",
			SystemPrompt: "You are a helpful, concise AI assistant.",
		},
		UI: UIConfig{
			Theme:            "default",
			RenderThrottleMs: 40,
			ShowMetrics:      true,
			ShowTimestamps:   false,
			FontSize:         2,
		},
		Session: SessionConfig{
			AutoSave:    false,
			SessionsDir: "~/.openllama/sessions",
			MaxSessions: 100,
		},
		Debug: false,
	}
}
