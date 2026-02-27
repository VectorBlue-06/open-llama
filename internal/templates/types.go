package templates

// Template defines the format for a prompt template.
type Template struct {
	Name            string   `json:"name"`
	SystemPrefix    string   `json:"system_prefix"`
	SystemSuffix    string   `json:"system_suffix"`
	UserPrefix      string   `json:"user_prefix"`
	UserSuffix      string   `json:"user_suffix"`
	AssistantPrefix string   `json:"assistant_prefix"`
	AssistantSuffix string   `json:"assistant_suffix"`
	StopTokens      []string `json:"stop_tokens"`
}
