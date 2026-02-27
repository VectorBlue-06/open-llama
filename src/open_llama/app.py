"""Textual TUI application for open-llama."""

from __future__ import annotations

from textual import on, work
from textual.app import App, ComposeResult
from textual.binding import Binding
from textual.containers import ScrollableContainer, Vertical
from textual.screen import Screen
from textual.widgets import (
    Button,
    Footer,
    Header,
    Input,
    Label,
    ListItem,
    ListView,
    Static,
)

from .ollama_client import OllamaClient


# ---------------------------------------------------------------------------
# Helper widgets
# ---------------------------------------------------------------------------


class ChatMessage(Static):
    """A single message bubble in the conversation."""

    DEFAULT_CSS = """
    ChatMessage {
        padding: 0 2;
        margin: 0 0 1 0;
    }
    ChatMessage.user {
        color: $accent;
        text-align: right;
    }
    ChatMessage.assistant {
        color: $text;
    }
    """

    def __init__(self, role: str, content: str) -> None:
        prefix = "You" if role == "user" else "AI"
        super().__init__(f"[bold]{prefix}:[/bold] {content}", classes=role)


# ---------------------------------------------------------------------------
# Model selection screen
# ---------------------------------------------------------------------------


class ModelScreen(Screen[str]):
    """Screen that lets the user pick an Ollama model."""

    BINDINGS = [Binding("ctrl+q", "app.quit", "Quit")]

    DEFAULT_CSS = """
    ModelScreen {
        align: center middle;
    }
    #model-box {
        width: 60;
        height: auto;
        border: round $accent;
        padding: 1 2;
    }
    #title {
        text-align: center;
        margin-bottom: 1;
    }
    #no-models {
        color: $error;
        text-align: center;
    }
    """

    def __init__(self, client: OllamaClient) -> None:
        super().__init__()
        self._client = client

    def compose(self) -> ComposeResult:
        with Vertical(id="model-box"):
            yield Label("open-llama", id="title")
            yield ListView(id="model-list")
            yield Label("", id="no-models")
        yield Footer()

    def on_mount(self) -> None:
        self._load_models()

    @work(exclusive=True)
    async def _load_models(self) -> None:
        list_view = self.query_one("#model-list", ListView)
        no_models = self.query_one("#no-models", Label)
        try:
            models = await self._client.list_models()
        except Exception as exc:  # noqa: BLE001
            no_models.update(f"Could not connect to Ollama: {exc}")
            return

        if not models:
            no_models.update("No models found. Run 'ollama pull <model>' first.")
            return

        for name in models:
            list_view.append(ListItem(Label(name), name=name))

        list_view.focus()

    @on(ListView.Selected)
    def _on_select(self, event: ListView.Selected) -> None:
        self.dismiss(event.item.name)


# ---------------------------------------------------------------------------
# Chat screen
# ---------------------------------------------------------------------------


class ChatScreen(Screen[None]):
    """Main chat screen."""

    BINDINGS = [
        Binding("ctrl+n", "new_chat", "New chat"),
        Binding("ctrl+m", "switch_model", "Switch model"),
        Binding("ctrl+q", "app.quit", "Quit"),
    ]

    DEFAULT_CSS = """
    ChatScreen {
        layout: vertical;
    }
    #history {
        height: 1fr;
        border: round $surface;
    }
    #input-row {
        height: auto;
        padding: 0 1;
    }
    #message-input {
        width: 1fr;
    }
    """

    def __init__(self, client: OllamaClient, model: str) -> None:
        super().__init__()
        self._client = client
        self._model = model
        self._messages: list[dict[str, str]] = []
        self._generating = False

    def compose(self) -> ComposeResult:
        yield Header(show_clock=True)
        with ScrollableContainer(id="history"):
            pass
        with Vertical(id="input-row"):
            yield Input(placeholder="Type a message…", id="message-input")
        yield Footer()

    def on_mount(self) -> None:
        self.title = f"open-llama  [{self._model}]"
        self.query_one("#message-input", Input).focus()

    # ------------------------------------------------------------------
    # Sending a message
    # ------------------------------------------------------------------

    @on(Input.Submitted, "#message-input")
    def _on_submit(self, event: Input.Submitted) -> None:
        text = event.value.strip()
        if not text or self._generating:
            return
        event.input.clear()
        self._send(text)

    @work(exclusive=True)
    async def _send(self, text: str) -> None:
        self._generating = True
        history = self.query_one("#history", ScrollableContainer)

        # Show user bubble
        self._messages.append({"role": "user", "content": text})
        history.mount(ChatMessage("user", text))
        history.scroll_end(animate=False)

        # Placeholder for assistant response
        assistant_widget = ChatMessage("assistant", "")
        history.mount(assistant_widget)

        accumulated = ""
        try:
            async for token in self._client.chat_stream(self._model, self._messages):
                accumulated += token
                assistant_widget.update(
                    f"[bold]AI:[/bold] {accumulated}"
                )
                history.scroll_end(animate=False)
        except Exception as exc:  # noqa: BLE001
            assistant_widget.update(f"[bold]Error:[/bold] {exc}")

        self._messages.append({"role": "assistant", "content": accumulated})
        self._generating = False
        self.query_one("#message-input", Input).focus()

    # ------------------------------------------------------------------
    # Actions
    # ------------------------------------------------------------------

    def action_new_chat(self) -> None:
        self._messages.clear()
        history = self.query_one("#history", ScrollableContainer)
        history.remove_children()

    def action_switch_model(self) -> None:
        self.app.push_screen(
            ModelScreen(self._client),
            callback=self._on_model_selected,
        )

    def _on_model_selected(self, model: str | None) -> None:
        if model:
            self._model = model
            self.title = f"open-llama  [{self._model}]"
            self.action_new_chat()


# ---------------------------------------------------------------------------
# Application
# ---------------------------------------------------------------------------


class OpenLlamaApp(App[None]):
    """The root Textual application."""

    TITLE = "open-llama"

    def __init__(self, base_url: str) -> None:
        super().__init__()
        self._client = OllamaClient(base_url)

    def on_mount(self) -> None:
        self.push_screen(ModelScreen(self._client), callback=self._on_model_selected)

    def _on_model_selected(self, model: str | None) -> None:
        if model is None:
            self.exit()
        else:
            self.push_screen(ChatScreen(self._client, model))
