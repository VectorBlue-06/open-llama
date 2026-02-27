package templates

// ChatML is the default template — works with most modern models.
var ChatML = Template{
	Name:            "ChatML",
	SystemPrefix:    "<|im_start|>system\n",
	SystemSuffix:    "<|im_end|>\n",
	UserPrefix:      "<|im_start|>user\n",
	UserSuffix:      "<|im_end|>\n",
	AssistantPrefix: "<|im_start|>assistant\n",
	AssistantSuffix: "<|im_end|>\n",
	StopTokens:      []string{"<|im_end|>"},
}

// Llama2 template for Llama 2 family models.
var Llama2 = Template{
	Name:            "Llama 2",
	SystemPrefix:    "<s>[INST] <<SYS>>\n",
	SystemSuffix:    "\n<</SYS>>\n\n",
	UserPrefix:      "",
	UserSuffix:      " [/INST] ",
	AssistantPrefix: "",
	AssistantSuffix: " </s><s>[INST] ",
	StopTokens:      []string{"</s>"},
}

// Llama3 template for Llama 3 family models.
var Llama3 = Template{
	Name:            "Llama 3",
	SystemPrefix:    "<|begin_of_text|><|start_header_id|>system<|end_header_id|>\n\n",
	SystemSuffix:    "<|eot_id|>",
	UserPrefix:      "<|start_header_id|>user<|end_header_id|>\n\n",
	UserSuffix:      "<|eot_id|>",
	AssistantPrefix: "<|start_header_id|>assistant<|end_header_id|>\n\n",
	AssistantSuffix: "<|eot_id|>",
	StopTokens:      []string{"<|eot_id|>"},
}

// Alpaca template for Alpaca-style models.
var Alpaca = Template{
	Name:            "Alpaca",
	SystemPrefix:    "",
	SystemSuffix:    "\n\n",
	UserPrefix:      "### Instruction:\n",
	UserSuffix:      "\n\n",
	AssistantPrefix: "### Response:\n",
	AssistantSuffix: "\n\n",
	StopTokens:      []string{"### Instruction:", "###"},
}

// Minimal is a simple fallback template with no special tokens.
var Minimal = Template{
	Name:            "Minimal",
	SystemPrefix:    "System: ",
	SystemSuffix:    "\n\n",
	UserPrefix:      "User: ",
	UserSuffix:      "\n",
	AssistantPrefix: "Assistant: ",
	AssistantSuffix: "\n",
	StopTokens:      []string{"User:", "System:"},
}

// BuiltinTemplates returns all built-in templates keyed by lowercase name.
var BuiltinTemplates = map[string]Template{
	"chatml":  ChatML,
	"llama2":  Llama2,
	"llama3":  Llama3,
	"alpaca":  Alpaca,
	"minimal": Minimal,
}
