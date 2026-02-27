<div align="center">

# 🦙 OpenLlama

**Local AI in your terminal. Fast, private, zero setup.**

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go&logoColor=white)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-blue?style=flat)]()

<br>

A polished TUI chat application powered by [llama.cpp](https://github.com/ggerganov/llama.cpp).  
Runs 100% locally — no internet, no telemetry, no cloud.

</div>

---

## ✨ Features

- 🚀 **Instant startup** — TUI ready in under 200ms
- 🔒 **Fully offline** — zero network requests, zero telemetry
- 🎯 **Zero config** — auto-detects hardware, picks optimal settings
- 💬 **Streaming responses** — smooth token-by-token output
- 🧠 **Smart context** — automatic sliding-window management
- 🎨 **Beautiful TUI** — colors, spinners, keyboard-driven interface
- 🔄 **Hot-swap models** — switch models without restarting
- ⚡ **GPU accelerated** — CUDA and Metal support out of the box

## 🚀 Quick Start

```bash
# 1. Place a GGUF model
mkdir -p ~/.openllama/models
cp your-model.gguf ~/.openllama/models/

# 2. Place llama-server alongside the binary

# 3. Run
./openllama
```

## ⌨️ Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Enter` | Send message |
| `Esc` | Cancel / Close |
| `Ctrl+N` | New chat |
| `Ctrl+M` | Switch model |
| `Ctrl+T` | Switch template |
| `Ctrl+Q` | Quit |

## 🔧 Build from Source

```bash
git clone https://github.com/VectorBlue-06/open-llama.git
cd open-llama
make build
```

## 📖 Documentation

For complete documentation including configuration, templates, architecture, and troubleshooting:

**→ [Full Documentation](DOCUMENTATION.md)**

Quick links:
- [Usage Guide](docs/USAGE.md)
- [Configuration Reference](docs/CONFIG.md)
- [Prompt Templates](docs/TEMPLATES.md)
- [Implementation Plan](PLAN.md)

## 📋 System Requirements

| Resource | Minimum | Recommended |
|----------|---------|-------------|
| RAM | 4 GB | 16 GB+ |
| CPU | 4 cores | 8+ cores |
| GPU | Optional | NVIDIA CUDA / Apple Metal |

## 🛡️ Privacy

- **Zero telemetry** — no data ever leaves your machine
- **No analytics** — no usage tracking of any kind
- **Localhost only** — server never exposed to network
- **No cloud** — everything runs locally

## 📄 License

[MIT](LICENSE) — Use it however you want.
