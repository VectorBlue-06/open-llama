<div align="center">

# 🦙 OpenLlama

**Local AI in your terminal. Fast, private, zero setup.**

[![Go](https://img.shields.io/badge/Go-1.24.13+-00ADD8?style=flat&logo=go&logoColor=white)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-blue?style=flat)]()

<br>

A polished TUI chat application powered by [llama.cpp](https://github.com/ggerganov/llama.cpp).  
Runs 100% locally — no internet, no telemetry, no cloud.

</div>

---

##  Features

-  **Instant startup** — TUI ready in under 200ms
-  **Fully offline** — zero network requests, zero telemetry
-  **Zero config** — auto-detects hardware, picks optimal settings
-  **Streaming responses** — smooth token-by-token output
-  **Smart context** — automatic sliding-window management
-  **Beautiful TUI** — colors, spinners, keyboard-driven interface
-  **Hot-swap models** — switch models without restarting
-  **GPU accelerated** — CUDA and Metal support out of the box
-  **Startup launcher** — centered prompt/search with device + model info
-  **Floating settings menu** — edit model path, llama.cpp path, font size, and generation parameters

##  Quick Start

```bash
# 1. Place a GGUF model
mkdir -p runtime/models
cp your-model.gguf runtime/models/

# 2. Place llama-server in runtime/llama.cpp/ (next to the binary)

# 3. Run
./openllama
```

## ⌨️ Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Enter` | Send message |
| `Shift+Enter` | New line in input |
| `Tab` | Open settings menu |
| `Esc` | Open settings menu / close overlay / cancel stream |
| `Ctrl+N` | New chat |
| `Ctrl+M` | Switch model |
| `Ctrl+T` | Switch template |
| `Ctrl+Q` | Quit |

##  Startup + Settings Flow

- On launch, OpenLlama shows a centered startup screen with:
	- available RAM
	- GPU status
	- currently selected model
	- a centered prompt/search input
- Press `Enter` on the startup prompt to begin chat mode (input moves to bottom panel).
- Press `Tab` (or `Esc` when not streaming) to open the floating settings menu.
- In settings, you can edit:
	- llama.cpp install path
	- models directory path
	- font size (terminal-friendly scale)
	- generation params (`temperature`, `top_p`, `top_k`, `repeat_penalty`, `max_tokens`)
	- selected model

##  Build from Source

```bash
git clone https://github.com/VectorBlue-06/open-llama.git
cd open-llama
make build
```

##  Documentation

For complete documentation including configuration, templates, architecture, and troubleshooting:

**→ [Full Documentation](DOCUMENTATION.md)**

Quick links:
- [Usage Guide](docs/USAGE.md)
- [Configuration Reference](docs/CONFIG.md)
- [Prompt Templates](docs/TEMPLATES.md)
- [Implementation Plan](PLAN.md)

##  System Requirements

| Resource | Minimum | Recommended |
|----------|---------|-------------|
| RAM | 4 GB | 16 GB+ |
| CPU | 4 cores | 8+ cores |
| GPU | Optional | NVIDIA CUDA / Apple Metal |

##  Privacy

- **Zero telemetry** — no data ever leaves your machine
- **No analytics** — no usage tracking of any kind
- **Localhost only** — server never exposed to network
- **No cloud** — everything runs locally

##  License

[MIT](LICENSE) — Use it however you want.
