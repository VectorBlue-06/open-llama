# OpenLlama — Complete Documentation

> Comprehensive reference for OpenLlama — the local AI terminal assistant.

---

## Table of Contents

- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [Prompt Templates](#prompt-templates)
- [Hardware Detection](#hardware-detection)
- [Context Management](#context-management)
- [Server Integration](#server-integration)
- [Metrics & Performance](#metrics--performance)
- [Session Management](#session-management)
- [Security & Privacy](#security--privacy)
- [Building from Source](#building-from-source)
- [Testing](#testing)
- [Troubleshooting](#troubleshooting)
- [Roadmap](#roadmap)

---

## Architecture

OpenLlama is a single-binary Go application that orchestrates a local LLM inference server.

```
┌─────────────────────────────────────────────────────┐
│                   OpenLlama Binary                   │
│                                                      │
│  ┌──────────┐  ┌───────────┐  ┌──────────────────┐  │
│  │  Config   │  │  Hardware  │  │  Server Manager  │  │
│  │  Manager  │  │  Detector  │  │  (llama.cpp)     │  │
│  └────┬──────┘  └─────┬─────┘  └────────┬─────────┘  │
│       │               │                  │             │
│       ▼               ▼                  ▼             │
│  ┌──────────────────────────────────────────────┐     │
│  │              App Controller                   │     │
│  └──────────┬───────────────────┬───────────────┘     │
│             │                   │                      │
│     ┌───────▼──────┐    ┌──────▼───────────┐          │
│     │  Chat Engine  │    │  Bubble Tea UI    │          │
│     │  ┌──────────┐ │    │  ┌─────────────┐ │          │
│     │  │ Context  │ │    │  │ Top Bar      │ │          │
│     │  │ Manager  │ │    │  │ Chat View    │ │          │
│     │  ├──────────┤ │    │  │ Input Box    │ │          │
│     │  │ Template │ │    │  │ Status Bar   │ │          │
│     │  │ Engine   │ │    │  └─────────────┘ │          │
│     │  └──────────┘ │    └──────────────────┘          │
│     └──────────────┘                                   │
│                                                        │
│       127.0.0.1:random_port                            │
│             │                                          │
│     ┌───────▼──────────────────────────┐               │
│     │  llama-server (child process)     │               │
│     │  - Loads GGUF model              │               │
│     │  - Serves /completion endpoint   │               │
│     │  - Bound to localhost only       │               │
│     └──────────────────────────────────┘               │
└────────────────────────────────────────────────────────┘
```

### Communication Flow

- **App → llama-server**: HTTP requests to `http://127.0.0.1:{port}`
- **Streaming**: Server-Sent Events (SSE) via `/completion` endpoint
- **Health Check**: `GET /health` with retry loop
- **Process Control**: `os/exec` with signal-based shutdown

---

## Project Structure

```
openllama/
├── cmd/openllama/main.go         # Entry point
├── internal/
│   ├── app/app.go                # App lifecycle controller
│   ├── ui/                       # Bubble Tea TUI
│   │   ├── model.go              # Root UI model
│   │   ├── update.go             # Message handling
│   │   ├── view.go               # Rendering
│   │   ├── keymap.go             # Key bindings
│   │   ├── styles.go             # Colors and styles
│   │   └── messages.go           # Custom messages
│   ├── chat/                     # Chat engine
│   │   ├── engine.go             # Conversation management
│   │   ├── message.go            # Message types
│   │   └── stream.go             # SSE streaming client
│   ├── context/                  # Context window
│   │   ├── manager.go            # Sliding window manager
│   │   └── tokenizer.go          # Token estimation
│   ├── server/                   # Server management
│   │   ├── server.go             # Lifecycle (start/stop)
│   │   ├── embed.go              # Binary discovery
│   │   └── port.go               # Port allocation
│   ├── templates/                # Prompt templates
│   │   ├── engine.go             # Template formatting
│   │   ├── builtin.go            # Built-in templates
│   │   └── types.go              # Template struct
│   ├── config/                   # Configuration
│   │   ├── config.go             # Config struct
│   │   ├── loader.go             # Load/save JSON
│   │   └── paths.go              # OS path resolution
│   ├── hardware/                 # Hardware detection
│   │   ├── detect.go             # Common detection
│   │   ├── detect_linux.go       # Linux (CUDA)
│   │   ├── detect_darwin.go      # macOS (Metal)
│   │   └── detect_windows.go     # Windows (CUDA)
│   ├── models/                   # Model management
│   │   ├── scanner.go            # GGUF discovery
│   │   └── info.go               # Model info struct
│   ├── metrics/collector.go      # Runtime metrics
│   └── utils/                    # Utilities
│       ├── logger.go             # Structured logging
│       └── fs.go                 # File system helpers
├── configs/default.json          # Default configuration
├── assets/server/                # Server binary location
├── scripts/                      # Build scripts
├── docs/                         # Documentation
├── Makefile                      # Build automation
├── LICENSE                       # MIT License
└── README.md
```

---

## Installation

### Pre-built Binaries

Download from [Releases](https://github.com/VectorBlue-06/open-llama/releases):

| Platform | File |
|----------|------|
| Linux x86_64 | `openllama-linux-amd64.tar.gz` |
| macOS ARM64 | `openllama-darwin-arm64.tar.gz` |
| Windows x86_64 | `openllama-windows-amd64.zip` |

### From Source

```bash
git clone https://github.com/VectorBlue-06/open-llama.git
cd open-llama
make build
```

### Requirements

| Resource | Minimum | Recommended |
|----------|---------|-------------|
| RAM | 4 GB | 16 GB+ |
| CPU | 4 cores | 8+ cores |
| Disk | 100 MB + model | — |
| GPU | Optional | NVIDIA CUDA 11.7+ / Apple Metal |
| OS | Linux x86_64, Windows 10+, macOS 12+ | — |

---

## Configuration

Config is stored at `~/.openllama/config.json`. Created automatically on first run.

See [docs/CONFIG.md](docs/CONFIG.md) for the complete reference.

### Key Settings

| Setting | Default | Description |
|---------|---------|-------------|
| `server.ctx_size` | 4096 | Context window (tokens) |
| `generation.temperature` | 0.7 | Response creativity |
| `template.default` | chatml | Prompt template |
| `session.auto_save` | false | Auto-save on quit |
| `debug` | false | Verbose logging |

### Priority Order

1. CLI flags (highest)
2. Environment variables
3. Config file
4. Defaults (lowest)

---

## Usage

### Quick Start

```bash
# 1. Place a model
mkdir -p ~/.openllama/models
cp your-model.gguf ~/.openllama/models/

# 2. Place llama-server next to the binary
cp llama-server /same/dir/as/openllama/

# 3. Run
./openllama
```

### Keyboard Shortcuts

| Key | Action | Context |
|-----|--------|---------|
| `Enter` | Send message | Input has text |
| `Esc` | Cancel stream / close overlay | During stream or overlay |
| `Ctrl+N` | New conversation | Always |
| `Ctrl+M` | Model picker | Not streaming |
| `Ctrl+T` | Template picker | Not streaming |
| `Ctrl+S` | Save session | Always |
| `Ctrl+Q` / `Ctrl+C` | Quit | Always |
| `↑/↓` | Scroll chat | Chat view |
| `PgUp/PgDn` | Fast scroll | Chat view |

### CLI Flags

```
--debug          Enable debug logging
--config PATH    Custom config file path
--model NAME     Model filename override
--port PORT      Server port (0 = auto)
--version        Print version
```

---

## Prompt Templates

Five built-in templates are included. See [docs/TEMPLATES.md](docs/TEMPLATES.md) for details.

| Template | Best For |
|----------|----------|
| ChatML | Most modern models (default) |
| Llama 2 | Meta Llama 2 family |
| Llama 3 | Meta Llama 3 family |
| Alpaca | Alpaca-style models |
| Minimal | Universal fallback |

Templates are auto-detected from the model filename when possible.

---

## Hardware Detection

OpenLlama automatically detects and configures:

| Component | Detection Method |
|-----------|-----------------|
| CPU cores | `runtime.NumCPU()` |
| RAM | `/proc/meminfo` (Linux), `sysctl` (macOS), WMI (Windows) |
| NVIDIA GPU | `nvidia-smi` |
| Apple Metal | `system_profiler` |

### Auto-Configuration

| Parameter | Rule |
|-----------|------|
| Threads | `min(CPU_CORES, 8)` |
| GPU layers | 999 if GPU detected, else 0 |
| Context size | Adjusted based on free RAM |
| Batch size | 512 (default) |

---

## Context Management

Uses a deterministic sliding-window approach:

- **Token estimation**: ~1 token per 3.6 characters (English)
- **Window management**: Newest messages kept, oldest trimmed
- **System prompt**: Always preserved
- **Response reserve**: 512 tokens reserved for generation

The context percentage is shown in the top bar and changes color:
- 🟢 Green: < 50%
- 🟡 Yellow: 50-80%
- 🔴 Red: > 80%

---

## Server Integration

OpenLlama manages llama-server as a child process:

1. Finds binary (sidecar → `~/.openllama/bin/` → PATH)
2. Launches on random localhost port (49152-65535)
3. Polls `/health` until ready (timeout: 120s)
4. Communicates via HTTP + SSE streaming
5. Graceful shutdown on exit (SIGTERM → wait 5s → SIGKILL)

### API Endpoints Used

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/health` | GET | Readiness check |
| `/completion` | POST | Streaming text generation |

---

## Metrics & Performance

Live metrics displayed in the top bar:

| Metric | Source |
|--------|--------|
| Tokens/sec | Server completion response |
| Context usage | Context manager |
| GPU status | Hardware detection |

### Performance Targets

| Metric | Target |
|--------|--------|
| Startup to TUI | < 200ms |
| UI render | < 16ms/frame |
| Memory overhead | < 20 MB |
| Streaming | Smooth up to 100 tok/s |

---

## Session Management

Sessions can be saved as JSON in `~/.openllama/sessions/`.

- **Auto-save**: Enable in config (`session.auto_save: true`)
- **Manual save**: `Ctrl+S`
- **Limit**: Configurable (default: 100 sessions)

---

## Security & Privacy

- **100% offline** — zero network requests after setup
- **Zero telemetry** — no analytics or tracking
- **Localhost only** — server binds to `127.0.0.1`
- **Secure file permissions** — config `0600`, dirs `0700`
- **Process isolation** — server killed when app exits

---

## Building from Source

### Requirements

- Go 1.22+
- Make (optional)
- llama-server binary

### Build Commands

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Run tests
make test

# Clean
make clean
```

### Cross-Compilation

```bash
# Linux
GOOS=linux GOARCH=amd64 make build-linux

# macOS
GOOS=darwin GOARCH=arm64 make build-darwin

# Windows
GOOS=windows GOARCH=amd64 make build-windows
```

---

## Testing

```bash
# All tests with race detector
make test

# Short tests only
make test-short

# Specific package
go test ./internal/context/... -v
```

### Test Coverage

| Package | Priority | Focus |
|---------|----------|-------|
| `context` | Critical | Token estimation, sliding window |
| `templates` | Critical | All template formats |
| `config` | High | Load/save, defaults, merge |
| `server` | High | Port, args, health check |
| `chat` | High | Messages, SSE parsing |
| `hardware` | Medium | Detection edge cases |
| `models` | Medium | GGUF scanning, parsing |
| `metrics` | Medium | Thread safety |

---

## Troubleshooting

| Problem | Solution |
|---------|----------|
| No models found | Place `.gguf` files in `~/.openllama/models/` |
| Server not found | Place `llama-server` next to binary or in `~/.openllama/bin/` |
| Model too slow | Try a smaller quantization (Q4 instead of Q8) |
| Out of memory | Use a smaller model or lower `ctx_size` |
| Server timeout | Check `~/.openllama/openllama.log` |
| Terminal too small | Minimum size: 40×10 |

### Debug Mode

```bash
./openllama --debug
```

Logs to both `~/.openllama/openllama.log` and stderr.

---

## Roadmap

### Future Features

| Feature | Description |
|---------|-------------|
| Session browser | Browse/reload past sessions |
| Model downloader | Download from HuggingFace |
| Full markdown | Syntax highlighting in responses |
| Multi-tab chats | Multiple conversations |
| Export | Markdown, text, HTML export |
| Vim keybindings | Optional vim-style navigation |

---

*For the original implementation plan, see [PLAN.md](PLAN.md).*
