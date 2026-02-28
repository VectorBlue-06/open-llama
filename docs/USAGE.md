# Usage Guide

## Getting Started

### 1. Download OpenLlama

Download the latest release for your platform from the [Releases](https://github.com/VectorBlue-06/open-llama/releases) page.

### 2. Get a Model

Download a GGUF model file from [HuggingFace](https://huggingface.co). Recommended starter models:

| Model | Size | RAM Needed | Quality |
|-------|------|-----------|---------|
| Phi-3 Mini Q4_K_M | ~2.4 GB | ~3 GB | Good for quick tasks |
| Mistral 7B Q4_K_M | ~4.4 GB | ~5 GB | Great all-around |
| Llama 3 8B Q4_K_M | ~5.0 GB | ~6 GB | Best quality |

### 3. Place the Model

```bash
mkdir -p runtime/models
cp your-model.gguf runtime/models/
```

### 4. Get llama-server

Download the llama-server binary from [llama.cpp releases](https://github.com/ggerganov/llama.cpp/releases) and place it:
- In `runtime/llama.cpp/` next to the `openllama` binary (recommended), or
- Next to the `openllama` binary, or
- In `~/.openllama/bin/`, or
- Anywhere in your `PATH`

### 5. Run

```bash
./openllama
```

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Enter` | Send message |
| `Shift+Enter` | New line in input |
| `Tab` | Open settings menu |
| `Esc` | Open settings menu / cancel stream / close overlay |
| `Ctrl+N` | New conversation |
| `Ctrl+M` | Open model picker |
| `Ctrl+T` | Open template picker |
| `Ctrl+S` | Save session |
| `Ctrl+Q` | Quit |
| `Ctrl+C` | Quit |
| `↑/↓` | Scroll chat |
| `PgUp/PgDn` | Fast scroll |

## Startup Screen

At startup, OpenLlama opens with a centered launcher screen that shows:

- available RAM
- GPU availability
- selected model
- a centered prompt/search input

Type your first prompt and press `Enter` to switch into normal chat mode (bottom input + chat history view).

## Settings Menu

Open the floating settings menu with `Tab` (or `Esc` when not streaming). Use `↑/↓` to move and `Enter` to edit/apply.

Available settings:

- llama.cpp path
- models path
- font size
- selected model
- `temperature`, `top_p`, `top_k`, `repeat_penalty`, `max_tokens`

## CLI Flags

```
--debug          Enable debug logging
--config PATH    Path to config file
--model NAME     Model filename to use
--port PORT      Port for llama-server (0 = auto)
--version        Print version and exit
```

## Model Switching

Press `Ctrl+M` to open the model picker. Use arrow keys to select a model and press `Enter`. The server will restart with the new model.

## Templates

OpenLlama includes built-in prompt templates:

- **ChatML** (default) — Works with most modern models
- **Llama 2** — For Llama 2 family models
- **Llama 3** — For Llama 3 family models
- **Alpaca** — For Alpaca-style models
- **Minimal** — Simple fallback

Press `Ctrl+T` to switch templates during a session.

## Session Management

- Sessions can be auto-saved on quit (enable `auto_save` in config)
- Manual save with `Ctrl+S`
- Sessions are stored in `~/.openllama/sessions/`

## Troubleshooting

### "No models found"
Place `.gguf` model files in `runtime/models/` next to the `openllama` binary.

### "llama-server not found"
Download the llama-server binary and place it in `runtime/llama.cpp/` next to the `openllama` binary.

### "Model is taking too long to load"
The model may be too large for your available RAM. Try a smaller model or a lower quantization (Q4 instead of Q8).

### "Server not responding"
Check `~/.openllama/openllama.log` for error details.
