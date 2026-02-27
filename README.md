# open-llama

A TUI-based Ollama chat client — fast, keyboard-driven, no extra pain.

## Requirements

* Python ≥ 3.10
* [Ollama](https://ollama.com/) running locally (`ollama serve`)

## Installation

```bash
pip install -e .
```

Or install dependencies directly:

```bash
pip install -r requirements.txt
```

## Usage

```bash
open-llama                        # connects to http://localhost:11434 by default
open-llama --host http://host:11434  # custom Ollama server
```

On launch you will see a list of your locally-pulled models.  
Select one to open the chat screen.

### Keyboard shortcuts

| Key | Action |
|-----|--------|
| `Enter` | Send message |
| `Ctrl+N` | Start a new conversation |
| `Ctrl+M` | Switch to a different model |
| `Ctrl+Q` / `Ctrl+C` | Quit |

## Project layout

```
src/open_llama/
├── __init__.py
├── main.py             CLI entry-point
├── app.py              Textual TUI (ModelScreen + ChatScreen)
└── ollama_client.py    Async HTTP client for the Ollama REST API
```

See [PLAN.md](PLAN.md) for the full architecture and design decisions.
