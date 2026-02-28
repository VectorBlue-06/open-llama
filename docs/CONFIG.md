# Configuration Reference

OpenLlama stores its configuration in `~/.openllama/config.json`.

## Config File Location

| OS | Path |
|----|------|
| Linux | `~/.openllama/config.json` |
| macOS | `~/.openllama/config.json` |
| Windows | `%USERPROFILE%\.openllama\config.json` |

## Full Schema

```json
{
    "version": 1,
    "model": {
        "default": "",
        "llama_path": "runtime/llama.cpp",
        "models_dir": "runtime/models"
    },
    "server": {
        "host": "127.0.0.1",
        "port": 0,
        "ctx_size": 4096,
        "batch_size": 512,
        "threads": 0,
        "gpu_layers": -1,
        "extra_args": []
    },
    "generation": {
        "temperature": 0.7,
        "top_p": 0.9,
        "top_k": 40,
        "repeat_penalty": 1.1,
        "max_tokens": 2048
    },
    "template": {
        "default": "chatml",
        "system_prompt": "You are a helpful, concise AI assistant."
    },
    "ui": {
        "theme": "default",
        "render_throttle_ms": 40,
        "show_metrics": true,
        "show_timestamps": false,
        "font_size": 2
    },
    "session": {
        "auto_save": false,
        "sessions_dir": "~/.openllama/sessions",
        "max_sessions": 100
    },
    "debug": false
}
```

## Field Reference

### Model

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `default` | string | `""` | Filename of default model. Empty = auto-select. |
| `llama_path` | string | `runtime/llama.cpp` | Path to `llama-server` binary or containing directory. |
| `models_dir` | string | `runtime/models` | Directory to scan for .gguf files. |

### Server

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `host` | string | `127.0.0.1` | Bind address (always localhost). |
| `port` | int | `0` | Port for server. `0` = random. |
| `ctx_size` | int | `4096` | Context window size in tokens. |
| `batch_size` | int | `512` | Batch size for prompt processing. |
| `threads` | int | `0` | CPU threads. `0` = auto-detect. |
| `gpu_layers` | int | `-1` | GPU layers. `-1` = auto. `0` = CPU only. |
| `extra_args` | []string | `[]` | Extra CLI args for llama-server. |

### Generation

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `temperature` | float | `0.7` | Sampling temperature (0.0-2.0). |
| `top_p` | float | `0.9` | Nucleus sampling threshold. |
| `top_k` | int | `40` | Top-k sampling. |
| `repeat_penalty` | float | `1.1` | Repetition penalty. |
| `max_tokens` | int | `2048` | Max tokens per response. |

### Template

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `default` | string | `"chatml"` | Template name (chatml, llama2, llama3, alpaca, minimal). |
| `system_prompt` | string | (see above) | System prompt for all conversations. |

### UI

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `theme` | string | `default` | Color theme preset. |
| `render_throttle_ms` | int | `40` | UI update throttle interval. |
| `show_metrics` | bool | `true` | Show runtime metrics in top bar. |
| `show_timestamps` | bool | `false` | Show message timestamps. |
| `font_size` | int | `2` | Terminal-friendly text scale used by chat rendering. |

### Session

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `auto_save` | bool | `false` | Auto-save sessions on quit. |
| `sessions_dir` | string | `~/.openllama/sessions` | Session storage directory. |
| `max_sessions` | int | `100` | Maximum saved sessions. |

## Priority Order

Configuration is applied in this order (highest priority first):

1. CLI flags (`--model`, `--port`, `--debug`)
2. Environment variables (`OPENLLAMA_MODEL`, `OPENLLAMA_PORT`)
3. Config file (`~/.openllama/config.json`)
4. Built-in defaults
