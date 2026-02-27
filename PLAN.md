# open-llama – Project Plan

## Goal
Build a fast, minimal TUI (terminal user interface) that lets users chat with local
LLMs through [Ollama](https://ollama.com/), without any browser or heavy GUI.

## Architecture

```
open-llama/
├── PLAN.md                      ← this file
├── README.md                    ← user-facing docs
├── pyproject.toml               ← package metadata & entry-points
├── requirements.txt             ← runtime dependencies
└── src/
    └── open_llama/
        ├── __init__.py
        ├── main.py              ← CLI entry-point  (`open-llama` command)
        ├── app.py               ← Textual TUI application
        └── ollama_client.py     ← async HTTP client for the Ollama REST API
```

## Features
1. **Model picker** – lists all locally-pulled Ollama models; user selects one to
   start a session.
2. **Chat screen** – scrollable conversation history with distinct user / assistant
   bubbles.
3. **Streaming** – tokens are printed incrementally as Ollama produces them.
4. **Keyboard-first** – no mouse required.
   * `Enter` – send message
   * `Ctrl+N` – new conversation
   * `Ctrl+M` – switch model
   * `Ctrl+Q` / `Ctrl+C` – quit

## Technology choices
| Concern | Library |
|---------|---------|
| TUI framework | [Textual](https://github.com/Textualize/textual) |
| HTTP / streaming | `httpx` (async) |
| Python version | ≥ 3.10 |

## Implementation steps
- [x] Create PLAN.md
- [x] Set up `pyproject.toml` and `requirements.txt`
- [x] Implement `ollama_client.py` – pull models list, streaming chat
- [x] Implement `app.py` – Textual screens: `ModelScreen`, `ChatScreen`
- [x] Implement `main.py` – parse args, launch app
- [x] Update `README.md` with installation and usage instructions
