# Prompt Templates

OpenLlama uses prompt templates to format conversations for different model families.

## Built-in Templates

### ChatML (Default)

Works with most modern models including Mistral, OpenHermes, and many others.

```
<|im_start|>system
You are a helpful assistant.<|im_end|>
<|im_start|>user
Hello!<|im_end|>
<|im_start|>assistant
```

### Llama 2

For Meta's Llama 2 family models.

```
<s>[INST] <<SYS>>
You are a helpful assistant.
<</SYS>>

Hello! [/INST]
```

### Llama 3

For Meta's Llama 3 family models.

```
<|begin_of_text|><|start_header_id|>system<|end_header_id|>

You are a helpful assistant.<|eot_id|><|start_header_id|>user<|end_header_id|>

Hello!<|eot_id|><|start_header_id|>assistant<|end_header_id|>

```

### Alpaca

For Stanford Alpaca-style instruction-tuned models.

```
You are a helpful assistant.

### Instruction:
Hello!

### Response:
```

### Minimal

Simple fallback with no special tokens. Works with any model.

```
System: You are a helpful assistant.

User: Hello!
Assistant:
```

## Custom Templates

Define a custom template in your `config.json`:

```json
{
    "template": {
        "custom_template": {
            "name": "My Custom",
            "system_prefix": "[SYSTEM] ",
            "system_suffix": "\n",
            "user_prefix": "[USER] ",
            "user_suffix": "\n",
            "assistant_prefix": "[BOT] ",
            "assistant_suffix": "\n",
            "stop_tokens": ["[USER]", "[SYSTEM]"]
        }
    }
}
```

## Auto-Detection

OpenLlama attempts to match the template to the model filename:

| Filename Contains | Template Selected |
|-------------------|-------------------|
| `chatml` | ChatML |
| `llama-2` or `llama2` | Llama 2 |
| `llama-3` or `llama3` | Llama 3 |
| `alpaca` | Alpaca |
| (anything else) | ChatML (default) |

You can always override this with `Ctrl+T` or the config file.
