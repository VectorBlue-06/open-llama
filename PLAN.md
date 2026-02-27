# OpenLlama — Final Implementation Plan

> **Local AI TUI Assistant**
> Single-binary, polished, offline-first terminal chat application
> Bundled llama.cpp server · Cross-platform (Linux priority, Windows & macOS)

---

## Table of Contents

1. [Product Vision](#1-product-vision)
2. [Technology Stack & Dependencies](#2-technology-stack--dependencies)
3. [Core Architecture](#3-core-architecture)
4. [Project Structure — Detailed](#4-project-structure--detailed)
5. [Runtime Flow — Detailed](#5-runtime-flow--detailed)
6. [Server Integration (llama.cpp)](#6-server-integration-llamacpp)
7. [Hardware Detection & Auto-Configuration](#7-hardware-detection--auto-configuration)
8. [Model Management](#8-model-management)
9. [Context Engine](#9-context-engine)
10. [Prompt Template Engine](#10-prompt-template-engine)
11. [Chat Engine](#11-chat-engine)
12. [UI Design (Bubble Tea) — Detailed](#12-ui-design-bubble-tea--detailed)
13. [Performance Strategy](#13-performance-strategy)
14. [Metrics & Stats Display](#14-metrics--stats-display)
15. [Config System — Detailed](#15-config-system--detailed)
16. [Error Handling — Detailed](#16-error-handling--detailed)
17. [Logging & Debug Mode](#17-logging--debug-mode)
18. [Session Persistence](#18-session-persistence)
19. [Security Model](#19-security-model)
20. [Build System & Packaging](#20-build-system--packaging)
21. [Testing Strategy](#21-testing-strategy)
22. [Polishing Layer](#22-polishing-layer)
23. [MVP Feature Set (Locked)](#23-mvp-feature-set-locked)
24. [Implementation Phases & Milestones](#24-implementation-phases--milestones)
25. [Phase 2 — Future Roadmap](#25-phase-2--future-roadmap)
26. [Design Principles](#26-design-principles)

---

## 1. Product Vision

### What We Are Building

A fast, minimal, fully-offline AI terminal assistant that:

- Runs 100% locally — no internet required after setup
- Starts instantly (< 200ms to TUI, excluding model load)
- Feels like ChatGPT in a terminal — streaming, responsive, polished
- Handles context intelligently with automatic sliding-window trimming
- Works on CPU and GPU (CUDA, Metal) seamlessly
- Requires zero technical setup — download, drop a model in, run

### Target User

Technical professionals (developers, sysadmins, data scientists) who:
- Want a private, offline AI assistant
- Are comfortable with the terminal but don't want to configure llama.cpp flags
- Need something that "just works" out of the box

### Non-Goals for MVP

- No built-in model downloader
- No RAG / embeddings / vector store
- No plugin system
- No tool/function calling
- No GUI / web interface
- No multi-user / network serving

---

## 2. Technology Stack & Dependencies

### Language

| Component | Technology |
|-----------|-----------|
| Application | **Go 1.22+** |
| TUI Framework | [Bubble Tea](https://github.com/charmbracelet/bubbletea) v0.25+ |
| TUI Layout | [Lip Gloss](https://github.com/charmbracelet/lipgloss) v0.10+ |
| Text Input | [Bubble Tea textarea](https://github.com/charmbracelet/bubbles) |
| Inference Backend | [llama.cpp server](https://github.com/ggerganov/llama.cpp) (bundled binary) |

### Go Module Dependencies

```
require (
    github.com/charmbracelet/bubbletea   v0.25+
    github.com/charmbracelet/lipgloss    v0.10+
    github.com/charmbracelet/bubbles     v0.18+
    github.com/shirou/gopsutil/v3        v3.24+    // hardware detection (CPU, RAM, GPU)
)
```

### External Dependencies (Bundled at Build Time)

| Dependency | Purpose | Source |
|-----------|---------|--------|
| `llama-server` | LLM inference server | Pre-built from llama.cpp releases |

Platform-specific binaries:
- `llama-server-linux-x86_64` — Linux AMD64
- `llama-server-linux-x86_64-cuda` — Linux AMD64 with CUDA
- `llama-server-darwin-arm64` — macOS Apple Silicon (Metal)
- `llama-server-darwin-x86_64` — macOS Intel
- `llama-server-windows-x86_64.exe` — Windows AMD64
- `llama-server-windows-x86_64-cuda.exe` — Windows AMD64 with CUDA

### System Requirements

| Resource | Minimum | Recommended |
|----------|---------|-------------|
| RAM | 4 GB (Q4 small models) | 16 GB+ |
| CPU | 4 cores | 8+ cores |
| Disk | 100 MB (app) + model size | — |
| GPU | Optional | NVIDIA (CUDA 11.7+) or Apple Metal |
| OS | Linux x86_64, Windows 10+, macOS 12+ | — |

---

## 3. Core Architecture

### High-Level Diagram

```
┌─────────────────────────────────────────────────────┐
│                   OpenLlama Binary                   │
│                                                     │
│  ┌─────────┐  ┌───────────┐  ┌──────────────────┐  │
│  │  Config  │  │  Hardware  │  │  Server Manager  │  │
│  │ Manager  │  │  Detector  │  │  (llama.cpp)     │  │
│  └────┬─────┘  └─────┬─────┘  └────────┬─────────┘  │
│       │              │                  │            │
│       ▼              ▼                  ▼            │
│  ┌─────────────────────────────────────────────┐    │
│  │              App Controller                  │    │
│  │  (orchestrates startup, lifecycle, shutdown) │    │
│  └──────────┬───────────────────┬──────────────┘    │
│             │                   │                    │
│     ┌───────▼──────┐    ┌──────▼──────────┐        │
│     │  Chat Engine  │    │  Bubble Tea UI   │        │
│     │  ┌──────────┐ │    │  ┌────────────┐ │        │
│     │  │ Context  │ │    │  │ Top Bar     │ │        │
│     │  │ Manager  │ │    │  │ Chat View   │ │        │
│     │  ├──────────┤ │    │  │ Input Box   │ │        │
│     │  │ Template │ │    │  │ Status Bar  │ │        │
│     │  │ Engine   │ │    │  └────────────┘ │        │
│     │  ├──────────┤ │    └─────────────────┘        │
│     │  │ HTTP     │ │                                │
│     │  │ Client   │ │                                │
│     │  └──────────┘ │                                │
│     └──────────────┘                                │
│                                                     │
│       127.0.0.1:random_port                          │
│             │                                        │
│     ┌───────▼──────────────────────────┐            │
│     │  llama-server (child process)     │            │
│     │  - Loads GGUF model              │            │
│     │  - Serves /completion endpoint   │            │
│     │  - Bound to localhost only       │            │
│     └──────────────────────────────────┘            │
└─────────────────────────────────────────────────────┘
```

### Communication Pattern

- **App → llama-server**: HTTP requests to `http://127.0.0.1:{port}`
- **Streaming**: Server-Sent Events (SSE) via `/completion` endpoint
- **Health Check**: GET `/health` with retry loop
- **Process Control**: `os/exec.Cmd` with `cmd.Process.Signal()` for shutdown

### Key Interfaces (Go)

```go
// internal/server/server.go
type Server interface {
    Start(cfg ServerConfig) error
    Stop() error
    Health() (bool, error)
    Port() int
}

// internal/chat/engine.go
type ChatEngine interface {
    Send(prompt string) (<-chan StreamToken, error)
    Reset()
    Messages() []Message
    SetTemplate(t Template)
    SetSystemPrompt(s string)
}

// internal/context/manager.go
type ContextManager interface {
    Add(msg Message)
    Build() string           // Returns full prompt with template applied
    TokenEstimate() int
    Trim(maxTokens int)
    Clear()
}

// internal/config/config.go
type ConfigManager interface {
    Load() (*Config, error)
    Save(cfg *Config) error
    Defaults() *Config
}

// internal/hardware/detect.go
type HardwareInfo struct {
    CPUCores    int
    TotalRAM    uint64   // bytes
    FreeRAM     uint64   // bytes
    HasCUDA     bool
    CUDAVersion string
    HasMetal    bool
    GPUName     string
    GPUVRAM     uint64   // bytes
}
```

---

## 4. Project Structure — Detailed

```
openllama/
├── cmd/
│   └── openllama/
│       └── main.go                 # Entry point: parse flags, init app, run
│
├── internal/
│   ├── app/
│   │   ├── app.go                  # App struct, lifecycle (Init -> Run -> Shutdown)
│   │   └── app_test.go
│   │
│   ├── ui/
│   │   ├── model.go                # Bubble Tea root Model
│   │   ├── update.go               # Update function (message handling)
│   │   ├── view.go                 # View function (rendering)
│   │   ├── keymap.go               # Key bindings definition
│   │   ├── styles.go               # Lip Gloss styles (colors, borders, padding)
│   │   ├── components/
│   │   │   ├── topbar.go           # Top status bar component
│   │   │   ├── chatview.go         # Scrollable chat message list
│   │   │   ├── inputbox.go         # Multi-line text input area
│   │   │   ├── statusbar.go        # Bottom status / hint bar
│   │   │   ├── modelpicker.go      # Model selection overlay
│   │   │   ├── templatepicker.go   # Template selection overlay
│   │   │   ├── welcome.go          # First-run welcome screen
│   │   │   └── loading.go          # Loading/spinner overlay
│   │   ├── messages.go             # Custom Bubble Tea messages (StreamChunkMsg, etc.)
│   │   └── ui_test.go
│   │
│   ├── chat/
│   │   ├── engine.go               # Chat engine: manages conversation, calls server
│   │   ├── message.go              # Message struct (Role, Content, Timestamp)
│   │   ├── stream.go               # HTTP SSE streaming client
│   │   └── engine_test.go
│   │
│   ├── context/
│   │   ├── manager.go              # Context window manager
│   │   ├── tokenizer.go            # Simple token estimator (chars/4 heuristic)
│   │   └── manager_test.go
│   │
│   ├── server/
│   │   ├── server.go               # Server lifecycle (start, stop, health check)
│   │   ├── embed.go                # Binary extraction / discovery
│   │   ├── port.go                 # Random free port finder
│   │   └── server_test.go
│   │
│   ├── templates/
│   │   ├── engine.go               # Template engine: applies chat format
│   │   ├── builtin.go              # Built-in templates (ChatML, Llama2, Alpaca, etc.)
│   │   ├── types.go                # Template struct definition
│   │   └── engine_test.go
│   │
│   ├── config/
│   │   ├── config.go               # Config struct and defaults
│   │   ├── loader.go               # Load / save JSON config file
│   │   ├── paths.go                # OS-specific path resolution (~/.openllama/)
│   │   └── config_test.go
│   │
│   ├── metrics/
│   │   ├── collector.go            # Collects tokens/sec, context usage, RAM, etc.
│   │   └── collector_test.go
│   │
│   ├── hardware/
│   │   ├── detect.go               # CPU, RAM, GPU detection
│   │   ├── detect_linux.go         # Linux-specific (CUDA via nvidia-smi)
│   │   ├── detect_darwin.go        # macOS-specific (Metal via system_profiler)
│   │   ├── detect_windows.go       # Windows-specific (CUDA via nvidia-smi)
│   │   └── detect_test.go
│   │
│   ├── models/
│   │   ├── scanner.go              # Scan models dir, parse GGUF metadata
│   │   ├── info.go                 # ModelInfo struct (name, size, quant, RAM estimate)
│   │   └── scanner_test.go
│   │
│   └── utils/
│       ├── logger.go               # Structured logger (file + optional stderr in debug)
│       └── fs.go                   # File system helpers (ensure dir, temp dir, etc.)
│
├── assets/
│   └── server/                     # llama-server binaries (one per platform, added at build)
│       ├── .gitkeep
│       └── README.md               # Instructions for placing server binaries
│
├── configs/
│   └── default.json                # Default config shipped with the app
│
├── scripts/
│   ├── build.sh                    # Cross-platform build script
│   ├── build.ps1                   # Windows build script (PowerShell)
│   ├── download-server.sh          # Download llama-server binaries from releases
│   └── package.sh                  # Create distributable archives
│
├── docs/
│   ├── USAGE.md                    # User guide
│   ├── CONFIG.md                   # Config reference
│   └── TEMPLATES.md                # Template format documentation
│
├── go.mod
├── go.sum
├── Makefile                        # Build targets (build, test, lint, clean, package)
├── LICENSE
├── README.md
├── PLAN.md                         # This file
└── .goreleaser.yml                 # Optional: GoReleaser config for automated releases
```

---

## 5. Runtime Flow — Detailed

### 5.1 Startup Sequence

```
main.go
  │
  ├─ 1. Parse CLI flags (--debug, --config, --model, --port)
  │
  ├─ 2. Initialize logger
  │     └─ If --debug: log to stderr + file
  │     └─ Else: log to file only (~/.openllama/openllama.log)
  │
  ├─ 3. Load config
  │     ├─ Check ~/.openllama/config.json
  │     ├─ If not found -> create with defaults
  │     └─ Merge CLI overrides (--model, --port override config)
  │
  ├─ 4. Ensure directories exist
  │     ├─ ~/.openllama/
  │     ├─ ~/.openllama/models/
  │     ├─ ~/.openllama/sessions/
  │     └─ ~/.openllama/tmp/
  │
  ├─ 5. Scan models directory
  │     ├─ Find all *.gguf files
  │     ├─ Parse GGUF header for metadata (quant type, parameter count)
  │     ├─ Estimate RAM usage per model
  │     └─ If no models found -> show welcome screen with instructions
  │
  ├─ 6. Detect hardware
  │     ├─ CPU: core count via runtime.NumCPU()
  │     ├─ RAM: total and free via gopsutil
  │     ├─ GPU: attempt nvidia-smi (CUDA) or system_profiler (Metal)
  │     └─ Build HardwareInfo struct
  │
  ├─ 7. Auto-configure server parameters
  │     ├─ threads = min(cpu_cores, 8)  [cap for efficiency]
  │     ├─ gpu_layers = 999 if GPU detected, else 0
  │     ├─ ctx_size = min(config.ctx_size, RAM-safe limit)
  │     └─ Apply any user overrides from config
  │
  ├─ 8. Locate llama-server binary
  │     ├─ Check alongside app binary (sidecar mode)
  │     ├─ Check in ~/.openllama/bin/
  │     ├─ Verify binary executes (--version)
  │     └─ If not found -> show error with download instructions
  │
  ├─ 9. Select model
  │     ├─ If config.default_model set and exists -> use it
  │     ├─ If only one model -> auto-select
  │     └─ If multiple -> show picker (first run)
  │
  ├─ 10. Start llama-server
  │      ├─ Find random free port (49152-65535)
  │      ├─ Launch via os/exec with args
  │      ├─ Redirect stdout/stderr to log file
  │      └─ Store *exec.Cmd and PID
  │
  ├─ 11. Wait for server ready
  │      ├─ Poll GET http://127.0.0.1:{port}/health
  │      ├─ Retry every 200ms
  │      ├─ Timeout after 120s (model loading can be slow)
  │      ├─ Show "Loading model..." spinner in TUI during wait
  │      └─ On failure -> show error, offer retry or model switch
  │
  └─ 12. Launch Bubble Tea TUI
        ├─ Initialize root model with all dependencies
        ├─ Start Bubble Tea program
        └─ Block until program exits
```

### 5.2 Chat Flow (Per Message)

```
User presses Enter
  │
  ├─ 1. Read input text from textarea
  ├─ 2. Trim whitespace; ignore if empty
  ├─ 3. Create Message{Role: "user", Content: text, Time: now}
  ├─ 4. Append to conversation history
  │
  ├─ 5. Context Manager builds prompt
  │     ├─ Apply template (system + all messages formatted)
  │     ├─ Estimate total tokens
  │     ├─ If over limit -> trim oldest user/assistant pairs
  │     ├─ Re-estimate after trim
  │     └─ Return final prompt string
  │
  ├─ 6. Send HTTP POST to /completion
  │     ├─ Body: {"prompt": built_prompt, "stream": true, "temperature": T, ...}
  │     ├─ Set "Accept: text/event-stream"
  │     └─ Open persistent connection
  │
  ├─ 7. Stream response tokens
  │     ├─ Read SSE data events
  │     ├─ Parse JSON: {"content": "...", "stop": false}
  │     ├─ Send StreamChunkMsg to Bubble Tea
  │     ├─ Throttle UI updates (accumulate for 40ms before triggering redraw)
  │     └─ On stop=true -> send StreamDoneMsg
  │
  ├─ 8. On StreamDoneMsg
  │     ├─ Create Message{Role: "assistant", Content: full_response}
  │     ├─ Append to conversation history
  │     ├─ Update metrics (tokens/sec, total tokens)
  │     ├─ Update context usage display
  │     └─ Re-enable input
  │
  └─ 9. Error path
        ├─ HTTP error -> show inline error message
        ├─ Timeout -> show "Server not responding"
        └─ Connection lost -> attempt auto-reconnect once
```

### 5.3 Shutdown Sequence

```
User presses Ctrl+Q (or Ctrl+C)
  │
  ├─ 1. Cancel any in-flight HTTP request
  ├─ 2. Save session if config.auto_save_sessions == true
  │     └─ Write to ~/.openllama/sessions/{timestamp}.json
  ├─ 3. Stop llama-server
  │     ├─ Send SIGTERM (Linux/macOS) or taskkill (Windows)
  │     ├─ Wait up to 5 seconds
  │     └─ If still running -> SIGKILL / force kill
  ├─ 4. Clean temp files
  │     └─ Remove any temp files from tmp/
  └─ 5. Exit with code 0
```

---

## 6. Server Integration (llama.cpp)

### Binary Management

**Sidecar approach** — the llama-server binary ships alongside the app binary:

```
openllama/
├── openllama          <- app binary
├── llama-server       <- server binary (same directory)
└── models/
```

At runtime, the app locates the server binary by:
1. Checking the directory of the running executable
2. Falling back to `~/.openllama/bin/llama-server`
3. Falling back to `llama-server` in PATH

### Server Launch Configuration

```go
type ServerConfig struct {
    BinaryPath  string   // Path to llama-server binary
    ModelPath   string   // Absolute path to .gguf model file
    Host        string   // Always "127.0.0.1"
    Port        int      // Random free port (49152-65535)
    CtxSize     int      // Context window size in tokens
    Threads     int      // Number of CPU threads
    GPULayers   int      // Number of layers to offload to GPU (0 = CPU only)
    BatchSize   int      // Batch size for prompt processing (default: 512)
    ExtraArgs   []string // Any additional user-specified flags
}
```

### Server Command Construction

```go
func (s *Server) buildArgs(cfg ServerConfig) []string {
    args := []string{
        "-m", cfg.ModelPath,
        "--host", cfg.Host,
        "--port", strconv.Itoa(cfg.Port),
        "--ctx-size", strconv.Itoa(cfg.CtxSize),
        "--threads", strconv.Itoa(cfg.Threads),
        "--batch-size", strconv.Itoa(cfg.BatchSize),
    }
    if cfg.GPULayers > 0 {
        args = append(args, "--n-gpu-layers", strconv.Itoa(cfg.GPULayers))
    }
    args = append(args, cfg.ExtraArgs...)
    return args
}
```

### Health Check

```go
func (s *Server) waitForReady(timeout time.Duration) error {
    deadline := time.Now().Add(timeout)
    client := &http.Client{Timeout: 2 * time.Second}

    for time.Now().Before(deadline) {
        resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/health", s.port))
        if err == nil && resp.StatusCode == 200 {
            resp.Body.Close()
            return nil
        }
        time.Sleep(200 * time.Millisecond)
    }
    return fmt.Errorf("server did not become ready within %v", timeout)
}
```

### API Endpoints Used

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/health` | GET | Server readiness check |
| `/completion` | POST | Text completion with streaming |
| `/v1/models` | GET | Loaded model info (optional) |

### Completion Request Format

```json
{
    "prompt": "<full formatted prompt>",
    "stream": true,
    "temperature": 0.7,
    "top_p": 0.9,
    "top_k": 40,
    "repeat_penalty": 1.1,
    "n_predict": 2048,
    "stop": ["<|im_end|>", "</s>"]
}
```

### Streaming Response Format (SSE)

```
data: {"content": "Hello", "stop": false}
data: {"content": " world", "stop": false}
data: {"content": "", "stop": true, "timings": {"predicted_per_second": 24.5, ...}}
```

---

## 7. Hardware Detection & Auto-Configuration

### Detection Strategy

```go
func Detect() (*HardwareInfo, error) {
    info := &HardwareInfo{}

    // CPU — always available
    info.CPUCores = runtime.NumCPU()

    // RAM — via gopsutil
    vmStat, err := mem.VirtualMemory()
    if err == nil {
        info.TotalRAM = vmStat.Total
        info.FreeRAM = vmStat.Available
    }

    // GPU — platform-specific (see detect_{os}.go)
    detectGPU(info)

    return info, nil
}
```

### GPU Detection — Linux/Windows (CUDA)

```go
// Run: nvidia-smi --query-gpu=name,memory.total,driver_version --format=csv,noheader
// Parse output: "NVIDIA GeForce RTX 4090, 24564 MiB, 535.129.03"
func detectCUDA(info *HardwareInfo) {
    cmd := exec.Command("nvidia-smi",
        "--query-gpu=name,memory.total,driver_version",
        "--format=csv,noheader")
    output, err := cmd.Output()
    if err != nil {
        return // no CUDA GPU
    }
    info.HasCUDA = true
    // parse GPUName, GPUVRAM from CSV
}
```

### GPU Detection — macOS (Metal)

```go
// Run: system_profiler SPDisplaysDataType
// Parse for Metal support and VRAM
func detectMetal(info *HardwareInfo) {
    cmd := exec.Command("system_profiler", "SPDisplaysDataType")
    output, err := cmd.Output()
    if err != nil {
        return
    }
    if strings.Contains(string(output), "Metal") {
        info.HasMetal = true
    }
}
```

### Auto-Configuration Rules

| Parameter | Rule |
|-----------|------|
| `threads` | `min(CPU_CORES, 8)` — capped at 8 for diminishing returns |
| `gpu_layers` | `999` if CUDA/Metal detected (offload all), else `0` |
| `ctx_size` | Base `4096`. If free RAM > 16 GB: allow `8192`. If free RAM < 4 GB: cap at `2048`. |
| `batch_size` | `512` (default, good balance for prompt processing) |
| User override | Any value set in `config.json` overrides the auto-detected value |

---

## 8. Model Management

### Data Directory Layout

```
~/.openllama/
├── models/                    # User places .gguf files here
│   ├── mistral-7b-q4_k_m.gguf
│   └── llama-3-8b-q5_k_m.gguf
├── sessions/                  # Auto-saved chat sessions
├── bin/                       # Alternative location for llama-server
├── config.json
└── openllama.log
```

### First-Run Experience (No Models Found)

```
┌─────────────────────────────────────────────────┐
│          Welcome to OpenLlama!                   │
│                                                  │
│  No models found.                                │
│                                                  │
│  Place .gguf model files in:                     │
│  ~/.openllama/models/                            │
│                                                  │
│  Recommended starter models:                     │
│  - Mistral 7B Q4_K_M  (~4.4 GB RAM)            │
│  - Llama 3 8B Q4_K_M  (~5.0 GB RAM)            │
│  - Phi-3 Mini Q4_K_M  (~2.4 GB RAM)            │
│                                                  │
│  Download from: https://huggingface.co           │
│                                                  │
│  Press 'r' to rescan  |  'q' to quit            │
└─────────────────────────────────────────────────┘
```

### GGUF Metadata Parsing

Read the GGUF file header (first ~1 KB) to extract:

```go
type ModelInfo struct {
    Filename       string  // "mistral-7b-q4_k_m.gguf"
    FilePath       string  // Full absolute path
    FileSize       int64   // Bytes on disk
    QuantType      string  // "Q4_K_M", "Q5_K_S", etc. (from filename or header)
    ParameterCount string  // "7B", "13B" (parsed from filename heuristic)
    Architecture   string  // "llama", "mistral", "phi" (from GGUF metadata)
    ContextLength  int     // Trained max context (from GGUF metadata)
    RAMEstimate    uint64  // Estimated RAM in bytes
}
```

### RAM Estimation Heuristic

```go
func EstimateRAM(fileSize int64) uint64 {
    // Model weights + ~20% overhead for KV cache + runtime buffers
    return uint64(float64(fileSize) * 1.2)
}
```

### Model Switching Flow

1. User presses `Ctrl+M`
2. Show overlay with model list (name, size, quant, RAM estimate)
3. User selects with arrow keys + Enter
4. Show "Switching model..." spinner
5. Stop current llama-server process
6. Start new llama-server with selected model
7. Wait for health check (show loading progress)
8. Clear conversation history
9. Update top bar with new model name

---

## 9. Context Engine

### Overview

Maintains the conversation within token limits using a deterministic sliding-window approach. No embeddings, no RAG, no external memory.

### Token Estimation

Character-based heuristic (accurate within ~10% for English):

```go
func EstimateTokens(text string) int {
    // Average: 1 token ~ 4 characters for English
    // Slightly aggressive (3.6) for safety margin
    return int(math.Ceil(float64(len(text)) / 3.6))
}
```

### Sliding Window Algorithm

```go
type ContextManager struct {
    systemPrompt       string
    messages           []Message    // Full history
    maxTokens          int          // e.g., 4096
    reserveForResponse int          // e.g., 512
}

func (cm *ContextManager) Build(template Template) string {
    available := cm.maxTokens - cm.reserveForResponse

    // System prompt always included
    prompt := template.FormatSystem(cm.systemPrompt)
    used := EstimateTokens(prompt)

    // Walk messages from newest to oldest, include as many as fit
    var included []Message
    for i := len(cm.messages) - 1; i >= 0; i-- {
        formatted := template.FormatMessage(cm.messages[i])
        cost := EstimateTokens(formatted)
        if used+cost > available {
            break
        }
        included = append([]Message{cm.messages[i]}, included...)
        used += cost
    }

    // Build final prompt string
    for _, msg := range included {
        prompt += template.FormatMessage(msg)
    }
    prompt += template.AssistantPrefix()

    return prompt
}
```

### Context Stats Exposed to UI

| Stat | Type | Description |
|------|------|-------------|
| `TokensUsed` | int | Estimated tokens in current prompt |
| `TokensMax` | int | Configured context window size |
| `MessagesTotal` | int | Total messages in conversation history |
| `MessagesIncluded` | int | Messages fitting in current window |
| `PercentUsed` | float64 | `TokensUsed / TokensMax * 100` |

---

## 10. Prompt Template Engine

### Template Structure

```go
type Template struct {
    Name            string   // Display name: "ChatML", "Llama 2", etc.
    SystemPrefix    string   // Text before system prompt
    SystemSuffix    string   // Text after system prompt
    UserPrefix      string   // Text before user message
    UserSuffix      string   // Text after user message
    AssistantPrefix string   // Text before assistant response
    AssistantSuffix string   // Text after assistant response
    StopTokens      []string // Tokens that signal end of generation
}
```

### Built-in Templates

#### ChatML (default — works with most modern models)

```go
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
```

**Produces:**
```
<|im_start|>system
You are a helpful assistant.<|im_end|>
<|im_start|>user
Hello!<|im_end|>
<|im_start|>assistant
```

#### Llama 2

```go
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
```

#### Llama 3

```go
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
```

#### Alpaca

```go
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
```

#### Minimal (no special tokens — fallback)

```go
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
```

### Custom User Templates

Users can define custom templates in `config.json`:

```json
{
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
```

### Template Auto-Detection (Stretch Goal)

Attempt to match template based on model filename:
- Filename contains "chatml" → ChatML
- Filename contains "llama-2" or "llama2" → Llama 2
- Filename contains "llama-3" or "llama3" → Llama 3
- Filename contains "alpaca" → Alpaca
- Default → ChatML

---

## 11. Chat Engine

### Message Structure

```go
type Role string

const (
    RoleSystem    Role = "system"
    RoleUser      Role = "user"
    RoleAssistant Role = "assistant"
)

type Message struct {
    Role      Role      `json:"role"`
    Content   string    `json:"content"`
    Timestamp time.Time `json:"timestamp"`
}
```

### Stream Token

```go
type StreamToken struct {
    Content string
    Stop    bool
    Timings *Timings // present only on final token
}

type Timings struct {
    PredictedPerSecond float64 `json:"predicted_per_second"`
    PromptTokens       int     `json:"prompt_n"`
    PredictedTokens    int     `json:"predicted_n"`
    PromptMS           float64 `json:"prompt_ms"`
    PredictedMS        float64 `json:"predicted_ms"`
}
```

### HTTP Streaming Client

```go
func (e *Engine) streamCompletion(ctx context.Context, prompt string) (<-chan StreamToken, error) {
    ch := make(chan StreamToken, 64) // buffered channel

    body := CompletionRequest{
        Prompt:        prompt,
        Stream:        true,
        Temperature:   e.config.Temperature,
        TopP:          e.config.TopP,
        TopK:          e.config.TopK,
        RepeatPenalty: e.config.RepeatPenalty,
        NPredict:      e.config.MaxTokens,
        Stop:          e.template.StopTokens,
    }

    go func() {
        defer close(ch)

        req, _ := http.NewRequestWithContext(ctx, "POST",
            fmt.Sprintf("http://127.0.0.1:%d/completion", e.serverPort),
            marshalBody(body))
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("Accept", "text/event-stream")

        resp, err := e.client.Do(req)
        if err != nil {
            ch <- StreamToken{Content: "[Error: " + err.Error() + "]", Stop: true}
            return
        }
        defer resp.Body.Close()

        scanner := bufio.NewScanner(resp.Body)
        for scanner.Scan() {
            line := scanner.Text()
            if !strings.HasPrefix(line, "data: ") {
                continue
            }
            data := strings.TrimPrefix(line, "data: ")
            var token StreamToken
            json.Unmarshal([]byte(data), &token)
            ch <- token
            if token.Stop {
                return
            }
        }
    }()

    return ch, nil
}
```

### Cancellation

When user presses `Esc` during streaming:
1. Cancel the context (`ctx.Cancel()`)
2. HTTP request is aborted
3. Partial response is kept as the assistant message
4. Input is re-enabled

---

## 12. UI Design (Bubble Tea) — Detailed

### Screen Layout

```
┌────────────────────────────────────────────────────────────────┐
│ TOP BAR                                                        │
│ Model: mistral-7b-q4  │ Template: ChatML │ CTX: 62% │ 24 t/s │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│  CHAT VIEW (scrollable)                                        │
│                                                                │
│  You:                                                          │
│  Explain the difference between TCP and UDP                    │
│                                                                │
│  Assistant:                                                    │
│  TCP (Transmission Control Protocol) is a connection-oriented  │
│  protocol that ensures reliable, ordered delivery of data...   │
│                                                                │
│  You:                                                          │
│  Which is better for gaming?                                   │
│                                                                │
│  Assistant:                                                    │
│  UDP is generally preferred for gaming because...█             │
│  (streaming)                                                   │
│                                                                │
├────────────────────────────────────────────────────────────────┤
│ INPUT BOX                                                      │
│ > Type your message...                                         │
│                                                                │
├────────────────────────────────────────────────────────────────┤
│ STATUS BAR                                                     │
│ Ctrl+N New │ Ctrl+M Model │ Ctrl+T Template │ Ctrl+Q Quit     │
└────────────────────────────────────────────────────────────────┘
```

### Color Scheme

```go
var (
    ColorPrimary    = lipgloss.Color("#7C3AED")  // Purple accent
    ColorSecondary  = lipgloss.Color("#6B7280")  // Gray
    ColorUser       = lipgloss.Color("#3B82F6")  // Blue for user messages
    ColorAssistant  = lipgloss.Color("#10B981")  // Green for assistant
    ColorError      = lipgloss.Color("#EF4444")  // Red for errors
    ColorDim        = lipgloss.Color("#4B5563")  // Dimmed text
    ColorHighlight  = lipgloss.Color("#F59E0B")  // Yellow for highlights
)
```

### Bubble Tea Model Structure

```go
type Model struct {
    // Dependencies
    chatEngine     *chat.Engine
    contextManager *context.Manager
    metricsCollector *metrics.Collector

    // UI Components
    topBar         components.TopBar
    chatView       components.ChatView
    inputBox       components.InputBox
    statusBar      components.StatusBar

    // Overlays
    modelPicker    components.ModelPicker
    templatePicker components.TemplatePicker
    welcome        components.Welcome
    loading        components.Loading

    // State
    width, height  int          // Terminal size
    streaming      bool         // Currently streaming a response
    showOverlay    OverlayType  // Which overlay is visible (None, ModelPicker, etc.)
    err            error        // Current error to display

    // Streaming
    streamBuffer   strings.Builder  // Accumulates tokens during streaming
    lastRender     time.Time        // For throttling renders
}
```

### Key Bindings

| Key | Action | Context |
|-----|--------|---------|
| `Enter` | Send message | Input has text, not streaming |
| `Shift+Enter` | Newline in input | Always in input box |
| `Esc` | Cancel streaming / close overlay | During stream or overlay |
| `Ctrl+N` | New conversation | Always |
| `Ctrl+M` | Open model picker | Not streaming |
| `Ctrl+T` | Open template picker | Not streaming |
| `Ctrl+S` | Save session to file | Always |
| `Ctrl+Q` | Quit application | Always |
| `Ctrl+C` | Quit application | Always |
| `Ctrl+L` | Clear screen (redraw) | Always |
| `Up/Down` | Scroll chat history | In chat view |
| `PgUp/PgDn` | Scroll chat fast | In chat view |
| `Home` | Scroll to top | In chat view |
| `End` | Scroll to bottom | In chat view |
| `Tab` | Cycle focus (input <-> chat) | Always |

### Custom Bubble Tea Messages

```go
// Sent when a stream chunk arrives from the server
type StreamChunkMsg struct {
    Content string
}

// Sent when streaming is complete
type StreamDoneMsg struct {
    FullContent string
    Timings     *chat.Timings
}

// Sent when an error occurs during streaming
type StreamErrorMsg struct {
    Err error
}

// Sent when server health check completes
type ServerReadyMsg struct{}

// Sent when server fails to start
type ServerFailedMsg struct {
    Err error
}

// Sent on a timer to throttle UI redraws
type TickMsg time.Time

// Sent when model scan completes
type ModelsScanCompleteMsg struct {
    Models []models.ModelInfo
}
```

### Markdown-Lite Rendering

For MVP, support basic formatting in assistant responses:

| Feature | Rendering |
|---------|-----------|
| `**bold**` | Bold (lipgloss) |
| `*italic*` | Italic (if terminal supports) |
| `` `code` `` | Highlighted background |
| ```` ```code block``` ```` | Indented, dim background |
| `- list items` | Bullet character + indent |
| `1. numbered` | Number + indent |
| `# Headers` | Bold + color |

Full markdown rendering deferred to Phase 2.

---

## 13. Performance Strategy

### Streaming Throttle

```go
const renderThrottleInterval = 40 * time.Millisecond // ~25 FPS

func (m *Model) handleStreamChunk(msg StreamChunkMsg) {
    m.streamBuffer.WriteString(msg.Content)

    now := time.Now()
    if now.Sub(m.lastRender) >= renderThrottleInterval {
        // Flush buffer to UI
        m.chatView.AppendStreaming(m.streamBuffer.String())
        m.streamBuffer.Reset()
        m.lastRender = now
    }
    // Final flush happens on StreamDoneMsg
}
```

### HTTP Client Configuration

```go
var httpClient = &http.Client{
    Timeout: 0, // No timeout for streaming
    Transport: &http.Transport{
        MaxIdleConns:        1,
        MaxIdleConnsPerHost: 1,
        IdleConnTimeout:     120 * time.Second,
        DisableKeepAlives:   false,
    },
}
```

### Memory Management

- Pre-allocate chat view buffer: `make([]string, 0, 1024)`
- Reuse `strings.Builder` for stream accumulation
- Limit message history to 10,000 messages (beyond this, persist to disk)
- Use `sync.Pool` for temporary allocations in hot path

### Performance Targets

| Metric | Target |
|--------|--------|
| Startup to TUI visible | < 200ms (excluding model load) |
| UI render latency | < 16ms per frame |
| Memory overhead (no model) | < 20 MB |
| Streaming smoothness | No visible jank up to 100 tok/s |
| Context up to 8K tokens | No UI lag |
| Goroutine count | < 10 during idle |

### Profiling Commands (Development)

```bash
# CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Memory profile
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine dump
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

Enable pprof only in debug mode (`--debug` flag).

---

## 14. Metrics & Stats Display

### Metrics Collector

```go
type Collector struct {
    mu sync.RWMutex

    // Per-response metrics (updated on StreamDoneMsg)
    LastTokensPerSec   float64
    LastPromptTokens   int
    LastPredictedTokens int

    // Cumulative metrics
    TotalTokens        int
    TotalMessages      int

    // Context metrics (updated on each prompt build)
    ContextUsed        int     // tokens
    ContextMax         int     // tokens
    ContextPercent     float64

    // Hardware (set at startup)
    CPUCores           int
    RAMTotal           uint64
    RAMUsed            uint64
    GPUActive          bool
    GPULayers          int
}
```

### Top Bar Format

```
 mistral-7b-q4 │ ChatML │ CTX 62% (2534/4096) │ 24.5 t/s │ GPU
```

Breakdown:
- **Model name**: truncated to 20 chars
- **Template**: current template name
- **Context**: percentage + token counts
- **Speed**: tokens per second (from last response)
- **GPU**: shown if GPU layers > 0

### Data Sources

| Metric | Source |
|--------|--------|
| Tokens/sec | `timings.predicted_per_second` from completion response |
| Prompt tokens | `timings.prompt_n` from completion response |
| Context usage | Calculated by Context Manager |
| RAM usage | `gopsutil` (sampled every 5s) |
| GPU status | Hardware detection at startup |

---

## 15. Config System — Detailed

### Config File Location

| OS | Path |
|----|------|
| Linux | `~/.openllama/config.json` |
| macOS | `~/.openllama/config.json` |
| Windows | `%USERPROFILE%\.openllama\config.json` |

### Full Config Schema

```json
{
    "version": 1,

    "model": {
        "default": "",
        "models_dir": "~/.openllama/models"
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
        "system_prompt": "You are a helpful, concise AI assistant.",
        "custom_template": null
    },

    "ui": {
        "theme": "default",
        "render_throttle_ms": 40,
        "show_metrics": true,
        "show_timestamps": false
    },

    "session": {
        "auto_save": false,
        "sessions_dir": "~/.openllama/sessions",
        "max_sessions": 100
    },

    "debug": false
}
```

### Config Field Descriptions

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `model.default` | string | `""` | Filename of default model. Empty = auto-select or prompt. |
| `model.models_dir` | string | `~/.openllama/models` | Directory to scan for .gguf files. |
| `server.host` | string | `127.0.0.1` | Bind address for llama-server. Always localhost. |
| `server.port` | int | `0` | Port for llama-server. `0` = random free port. |
| `server.ctx_size` | int | `4096` | Context window size in tokens. |
| `server.batch_size` | int | `512` | Batch size for prompt processing. |
| `server.threads` | int | `0` | CPU threads. `0` = auto-detect. |
| `server.gpu_layers` | int | `-1` | GPU layers. `-1` = auto (all if GPU, 0 if not). `0` = force CPU. |
| `server.extra_args` | []string | `[]` | Additional CLI args passed to llama-server. |
| `generation.temperature` | float | `0.7` | Sampling temperature. |
| `generation.top_p` | float | `0.9` | Top-p (nucleus) sampling. |
| `generation.top_k` | int | `40` | Top-k sampling. |
| `generation.repeat_penalty` | float | `1.1` | Repetition penalty. |
| `generation.max_tokens` | int | `2048` | Max tokens to generate per response. |
| `template.default` | string | `"chatml"` | Default template name. |
| `template.system_prompt` | string | (see above) | System prompt prepended to every conversation. |
| `session.auto_save` | bool | `false` | Auto-save sessions on quit. |
| `debug` | bool | `false` | Enable debug logging and pprof. |

### Config Loading Priority (highest wins)

1. CLI flags (`--model`, `--port`, `--debug`, etc.)
2. Environment variables (`OPENLLAMA_MODEL`, `OPENLLAMA_PORT`, etc.)
3. Config file (`~/.openllama/config.json`)
4. Built-in defaults

---

## 16. Error Handling — Detailed

### Error Categories & Recovery

| Error | Detection | User Experience | Recovery |
|-------|-----------|----------------|----------|
| **No models found** | Model scan returns empty | Welcome screen with instructions | User adds models, presses 'r' to rescan |
| **Server binary not found** | Binary not at expected paths | Error screen: "llama-server not found" with path instructions | User places binary, restarts |
| **Server fails to start** | Process exits with non-zero code | Error screen with stderr output (last 10 lines) | Retry button, or model switch |
| **Server health timeout** | Health check exceeds 120s | "Model is taking too long to load. It may be too large for your RAM." | Retry or switch to smaller model |
| **Port conflict** | `bind: address already in use` in stderr | Transparent — auto-retry with new port | Automatic (up to 3 retries) |
| **OOM (out of memory)** | Server killed by OS (exit code 137) | "Model requires more RAM than available. Try a smaller quantization." | Model picker shown |
| **HTTP request failure** | Connection refused / timeout | Inline error in chat: "[Server error — retrying...]" | Auto-retry once, then show persistent error |
| **Streaming interrupted** | Connection reset during SSE | Keep partial response, show "[Response interrupted]" | User can resend |
| **GGUF parse error** | Invalid file header | Skip file in model list, log warning | Transparent to user |
| **Config parse error** | Invalid JSON | Log warning, use defaults | Auto-recover with defaults |
| **Terminal too small** | Width < 40 or height < 10 | "Terminal too small. Minimum: 40x10" | Resize terminal |

### Error Display Styles

- **Fatal errors** (no models, no server): Full-screen error with instructions
- **Recoverable errors** (HTTP failure, timeout): Inline message in chat view
- **Warnings** (high RAM usage, slow speed): Subtle indicator in top bar

### Panic Recovery

```go
func main() {
    defer func() {
        if r := recover(); r != nil {
            // Log panic with stack trace
            log.Printf("PANIC: %v\n%s", r, debug.Stack())
            // Attempt graceful server shutdown
            if server != nil {
                server.Stop()
            }
            fmt.Fprintf(os.Stderr, "OpenLlama crashed. See ~/.openllama/openllama.log for details.\n")
            os.Exit(1)
        }
    }()
    // ...
}
```

---

## 17. Logging & Debug Mode

### Log Levels

| Level | Usage |
|-------|-------|
| `ERROR` | Failures that affect functionality |
| `WARN` | Degraded behavior (fallbacks triggered) |
| `INFO` | Lifecycle events (startup, model loaded, shutdown) |
| `DEBUG` | Verbose detail (HTTP requests, token counts, timing) |

### Log Output

| Mode | Destination |
|------|-------------|
| Normal | `~/.openllama/openllama.log` (file only) |
| Debug (`--debug`) | File + stderr |

### Log Format

```
2026-02-27T10:30:00.000Z [INFO]  config loaded from ~/.openllama/config.json
2026-02-27T10:30:00.005Z [INFO]  hardware: 8 cores, 16384 MB RAM, CUDA (RTX 4090, 24GB)
2026-02-27T10:30:00.010Z [INFO]  models found: 2 (mistral-7b-q4_k_m.gguf, llama-3-8b-q5_k_m.gguf)
2026-02-27T10:30:00.012Z [INFO]  starting llama-server on port 52341
2026-02-27T10:30:02.500Z [INFO]  server ready (model loaded in 2.49s)
2026-02-27T10:30:02.502Z [INFO]  TUI launched
2026-02-27T10:35:10.100Z [DEBUG] completion request: 1287 estimated tokens, template=ChatML
2026-02-27T10:35:12.300Z [DEBUG] completion done: 156 tokens in 2.2s (70.9 t/s)
```

### Log Rotation

- Max log file size: 10 MB
- On exceeding: rename to `openllama.log.1`, start new file
- Keep at most 2 old log files

---

## 18. Session Persistence

### Session File Format

```json
{
    "version": 1,
    "created_at": "2026-02-27T10:30:00Z",
    "updated_at": "2026-02-27T10:45:00Z",
    "model": "mistral-7b-q4_k_m.gguf",
    "template": "chatml",
    "system_prompt": "You are a helpful assistant.",
    "messages": [
        {
            "role": "user",
            "content": "Hello!",
            "timestamp": "2026-02-27T10:31:00Z"
        },
        {
            "role": "assistant",
            "content": "Hello! How can I help you today?",
            "timestamp": "2026-02-27T10:31:02Z"
        }
    ],
    "stats": {
        "total_tokens": 234,
        "message_count": 4
    }
}
```

### Session Storage

```
~/.openllama/sessions/
├── 2026-02-27_103000.json
├── 2026-02-27_143000.json
└── ...
```

### Auto-Save Behavior

If `config.session.auto_save == true`:
- Save on `Ctrl+Q` (quit)
- Save on `Ctrl+N` (new chat — saves current before clearing)
- Save on `Ctrl+S` (manual save)

### Session Limits

- Max 100 saved sessions (configurable)
- Oldest sessions deleted when limit exceeded
- Max session file size: ~5 MB (approximately 50K messages)

---

## 19. Security Model

### Network Isolation

- llama-server **always** binds to `127.0.0.1` (IPv4 loopback only)
- Random high port (49152-65535) to avoid conflicts
- No option to bind to `0.0.0.0` or external interfaces
- No authentication needed (localhost only)

### Data Privacy

- **Zero telemetry** — no data leaves the machine, ever
- **No analytics** — no usage tracking
- **No network requests** — the app makes zero outbound connections
- All data stored locally in `~/.openllama/`

### File Permissions

- Config file: `0600` (owner read/write only)
- Sessions directory: `0700` (owner only)
- Log files: `0600`
- Server binary: `0755` (executable)

### Process Isolation

- Server runs as a child process (same user)
- Server is killed when parent exits (even on crash, via process group)
- No shared memory or IPC beyond HTTP

---

## 20. Build System & Packaging

### Build Requirements

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.22+ | Compile the application |
| Make | any | Build automation |
| llama-server | latest | Pre-built binary (per platform) |

### Makefile Targets

```makefile
.PHONY: build test clean lint package

VERSION := $(shell git describe --tags --always --dirty)
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

# Build for current platform
build:
	go build $(LDFLAGS) -o bin/openllama ./cmd/openllama

# Build for all platforms
build-all: build-linux build-darwin build-windows

build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/openllama-linux-amd64 ./cmd/openllama

build-darwin:
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/openllama-darwin-arm64 ./cmd/openllama

build-windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/openllama-windows-amd64.exe ./cmd/openllama

# Run tests
test:
	go test ./... -v -race -count=1

# Lint
lint:
	golangci-lint run ./...

# Clean build artifacts
clean:
	rm -rf bin/ dist/

# Package for distribution
package: build-all
	./scripts/package.sh
```

### Package Structure (Distribution)

```
openllama-v1.0.0-linux-amd64.tar.gz
├── openllama                  # Application binary
├── llama-server               # llama.cpp server binary
├── README.md                  # Quick start guide
└── LICENSE

openllama-v1.0.0-windows-amd64.zip
├── openllama.exe
├── llama-server.exe
├── README.md
└── LICENSE
```

### Build Script (`scripts/download-server.sh`)

```bash
#!/bin/bash
# Downloads the correct llama-server binary for the target platform
# from the llama.cpp GitHub releases

LLAMA_CPP_VERSION="b4567"  # Pin to a specific release
BASE_URL="https://github.com/ggerganov/llama.cpp/releases/download/${LLAMA_CPP_VERSION}"

case "$1" in
    linux-amd64)
        URL="${BASE_URL}/llama-server-linux-x86_64"
        ;;
    linux-amd64-cuda)
        URL="${BASE_URL}/llama-server-linux-x86_64-cuda"
        ;;
    darwin-arm64)
        URL="${BASE_URL}/llama-server-darwin-arm64"
        ;;
    windows-amd64)
        URL="${BASE_URL}/llama-server-windows-x86_64.exe"
        ;;
    *)
        echo "Usage: $0 {linux-amd64|linux-amd64-cuda|darwin-arm64|windows-amd64}"
        exit 1
        ;;
esac

mkdir -p assets/server
curl -L -o "assets/server/llama-server" "$URL"
chmod +x "assets/server/llama-server"
echo "Downloaded llama-server for $1"
```

### CI/CD (GitHub Actions — Recommended)

```yaml
# .github/workflows/release.yml
name: Release
on:
  push:
    tags: ['v*']
jobs:
  build:
    strategy:
      matrix:
        include:
          - os: ubuntu-latest
            goos: linux
            goarch: amd64
          - os: macos-latest
            goos: darwin
            goarch: arm64
          - os: windows-latest
            goos: windows
            goarch: amd64
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: scripts/download-server.sh ${{ matrix.goos }}-${{ matrix.goarch }}
      - run: GOOS=${{ matrix.goos }} GOARCH=${{ matrix.goarch }} make build
      - run: scripts/package.sh
      - uses: softprops/action-gh-release@v1
        with:
          files: dist/*
```

---

## 21. Testing Strategy

### Test Categories

| Category | Location | Tool | Coverage Target |
|----------|----------|------|----------------|
| Unit tests | `*_test.go` alongside code | `go test` | 80%+ for core packages |
| Integration tests | `internal/app/integration_test.go` | `go test -tags=integration` | Key flows |
| Manual testing | — | Human tester | Full UI, streaming, all keybinds |

### Unit Test Priorities

| Package | What to Test | Priority |
|---------|-------------|----------|
| `context` | Token estimation accuracy, sliding window correctness, edge cases (empty, single message, overflow) | **Critical** |
| `templates` | All built-in templates produce correct output, custom templates parse correctly | **Critical** |
| `config` | Load/save round-trip, defaults applied, merge priority, invalid JSON handling | **High** |
| `server` | Port finder, arg builder, health check retry logic (mock server) | **High** |
| `chat` | Message management, SSE parsing (with mock HTTP server) | **High** |
| `hardware` | Mock command outputs, edge cases (no GPU, multiple GPUs) | **Medium** |
| `models` | GGUF scanning, filename parsing, RAM estimation | **Medium** |
| `metrics` | Collector accumulation, thread safety | **Medium** |

### Integration Tests

```go
// internal/app/integration_test.go
// Requires: llama-server binary and a small test model

func TestFullChatFlow(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    // 1. Start server with tiny test model
    // 2. Send a prompt
    // 3. Verify streaming response
    // 4. Verify context management
    // 5. Stop server
}
```

### Test Model

For integration tests, use a tiny model:
- `tinyllamas-stories-260k-q8_0.gguf` (~500 KB) — produces gibberish but tests the pipeline

---

## 22. Polishing Layer

### First-Run Experience

1. App detects no config → creates default config
2. App detects no models → shows welcome screen
3. Welcome screen has clear instructions + recommended models
4. After models are added → auto-scan and proceed

### Loading States

| State | Visual |
|-------|--------|
| App starting | Centered spinner: "Starting OpenLlama..." |
| Model loading | Centered spinner: "Loading model... (this may take a moment)" |
| Switching models | Overlay spinner: "Switching to {model}..." |
| Waiting for response | Blinking cursor in assistant message area |

### Spinner Implementation

Use Bubble Tea's built-in spinner (`bubbles/spinner`):
```go
spinner.New(
    spinner.WithSpinner(spinner.Dot),
    spinner.WithStyle(lipgloss.NewStyle().Foreground(ColorPrimary)),
)
```

### Visual Polish Checklist

- [x] Consistent color scheme across all components
- [x] Proper padding and borders on all panels
- [x] User messages visually distinct from assistant messages
- [x] Smooth auto-scroll during streaming
- [x] Context percentage changes color: green (< 50%), yellow (50-80%), red (> 80%)
- [x] Error messages are red with clear text
- [x] Keyboard hints always visible in status bar
- [x] Terminal resize handled gracefully (all components reflow)
- [x] No flickering or visual artifacts during streaming
- [x] Clean exit — terminal restored to normal state

---

## 23. MVP Feature Set (Locked)

These features are **in scope** for the first release:

| # | Feature | Status |
|---|---------|--------|
| 1 | Bundled llama-server (sidecar) | Required |
| 2 | Auto hardware detection (CPU, RAM, GPU) | Required |
| 3 | Auto server configuration | Required |
| 4 | Chat interface (scrollable, keyboard-driven) | Required |
| 5 | Streaming token display | Required |
| 6 | Context manager (sliding window) | Required |
| 7 | Prompt templates (ChatML, Llama2, Llama3, Alpaca, Minimal) | Required |
| 8 | Custom user template support | Required |
| 9 | Model scanning and selection | Required |
| 10 | Model switching (hot-swap with server restart) | Required |
| 11 | Config file (JSON) | Required |
| 12 | Live metrics bar (tokens/sec, context %, model name) | Required |
| 13 | Graceful shutdown | Required |
| 14 | Session save/load | Required |
| 15 | Welcome screen (first run) | Required |
| 16 | Error handling with recovery | Required |
| 17 | Debug logging mode | Required |
| 18 | Cross-platform support (Linux, Windows, macOS) | Required |

### Explicitly NOT in MVP

- Model downloader
- Session history browser
- Plugin system
- Tool/function calling
- RAG / embeddings
- Full markdown rendering
- Benchmark mode
- Web UI
- Multi-model conversations
- Image generation / multimodal

---

## 24. Implementation Phases & Milestones

### Phase 0: Project Setup (Day 1)

- [x] Initialize Go module
- [x] Set up project structure (all directories)
- [x] Create `go.mod` with dependencies
- [x] Create Makefile
- [x] Write README.md stub

### Phase 1: Foundation (Days 2-4)

- [ ] `internal/config` — Config struct, loader, defaults, path resolution
- [ ] `internal/hardware` — CPU/RAM detection, GPU detection stubs
- [ ] `internal/utils` — Logger, filesystem helpers
- [ ] `internal/server` — Port finder, server start/stop/health
- [ ] CLI flag parsing in `main.go`
- [ ] **Milestone**: App starts, loads config, launches and stops llama-server

### Phase 2: Chat Engine (Days 5-7)

- [ ] `internal/templates` — Template struct, all built-in templates
- [ ] `internal/context` — Token estimator, sliding window manager
- [ ] `internal/chat` — Message types, SSE streaming client, chat engine
- [ ] **Milestone**: Can send prompts and receive streaming responses (CLI/log output)

### Phase 3: TUI (Days 8-12)

- [ ] `internal/ui/model.go` — Root Bubble Tea model
- [ ] `internal/ui/components/topbar.go` — Status bar
- [ ] `internal/ui/components/chatview.go` — Scrollable chat
- [ ] `internal/ui/components/inputbox.go` — Text input
- [ ] `internal/ui/components/statusbar.go` — Key hints
- [ ] `internal/ui/styles.go` — Color scheme
- [ ] `internal/ui/keymap.go` — Key bindings
- [ ] Wire streaming into TUI with throttling
- [ ] **Milestone**: Full working chat in TUI with streaming

### Phase 4: Model Management (Days 13-14)

- [ ] `internal/models` — GGUF scanner, model info
- [ ] `internal/ui/components/modelpicker.go` — Model selector overlay
- [ ] `internal/ui/components/templatepicker.go` — Template selector overlay
- [ ] Model switching (server restart flow)
- [ ] **Milestone**: Can scan, select, and switch models

### Phase 5: Metrics & Polish (Days 15-17)

- [ ] `internal/metrics` — Metrics collector
- [ ] Wire metrics into top bar
- [ ] Loading spinners and states
- [ ] Welcome screen
- [ ] Session save/load
- [ ] **Milestone**: Complete polished UI with metrics

### Phase 6: Error Handling & Testing (Days 18-20)

- [ ] Comprehensive error handling (all error table entries)
- [ ] Unit tests for all core packages
- [ ] Integration test with test model
- [ ] Edge case testing (no models, bad config, server crashes)
- [ ] **Milestone**: Robust error handling, 80%+ test coverage

### Phase 7: Packaging & Release (Days 21-23)

- [ ] Build scripts for all platforms
- [ ] Server binary download script
- [ ] Package creation (tar.gz, zip)
- [ ] GitHub Actions CI/CD
- [ ] Final README and documentation
- [ ] **Milestone**: Distributable packages for Linux, macOS, Windows

---

## 25. Phase 2 — Future Roadmap

These features are planned for after MVP:

| Feature | Description | Complexity |
|---------|-------------|------------|
| Session history browser | Browse and reload past sessions from TUI | Medium |
| Model downloader | Download models from HuggingFace directly | High |
| Plugin system | Extend functionality via Go plugins or scripts | High |
| Tool calling | Allow model to call defined tools (shell, web, etc.) | High |
| File-based RAG | Load files into context for Q&A | High |
| Full markdown renderer | Complete markdown rendering with syntax highlighting | Medium |
| Benchmark mode | Measure and display detailed performance metrics | Low |
| Multi-conversation tabs | Multiple chats open simultaneously | Medium |
| Conversation export | Export to markdown, text, or HTML | Low |
| System prompt library | Pre-built system prompts for common tasks | Low |
| Vim key bindings | Optional vim-style navigation | Low |

---

## 26. Design Principles

### Core Principles (Non-Negotiable)

1. **Zero manual setup** — Download, add model, run. No flags, no config editing required.
2. **Fast** — Every interaction should feel instant. Streaming should be smooth.
3. **Deterministic** — Same input, same config → same behavior. No hidden state.
4. **Fully offline** — Zero network requests. No telemetry. No data exfiltration.
5. **Minimal RAM overhead** — The app itself should use < 20 MB. The model is the user's choice.
6. **No background services** — Nothing runs when the app is closed.
7. **Clean logs in debug mode** — When something goes wrong, the logs tell the full story.

### Code Principles

1. **No unnecessary abstraction** — If a function is used once, it doesn't need an interface.
2. **Explicit over implicit** — No magic. Configuration > convention where it matters.
3. **Fail loudly, recover gracefully** — Log every error, but show the user a clean message.
4. **Test the critical path** — Context management and template formatting must be bulletproof.
5. **No goroutine leaks** — Every goroutine must have a clear lifecycle and cancellation path.

### UX Principles

1. **Keyboard-first** — Every action reachable via keyboard. Mouse optional.
2. **Progressive disclosure** — Show essentials by default, details on demand.
3. **No surprises** — App does exactly what the user expects, nothing more.
4. **Helpful errors** — Every error message tells the user what happened and what to do.

---

*This is the complete implementation plan for OpenLlama v1.0. All decisions are final for MVP scope. Implementation begins with Phase 0 (project setup) and proceeds sequentially through Phase 7 (packaging).*
