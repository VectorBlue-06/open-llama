"""CLI entry-point for open-llama."""

from __future__ import annotations

import argparse

from .app import OpenLlamaApp
from .ollama_client import DEFAULT_BASE_URL


def main() -> None:
    parser = argparse.ArgumentParser(
        prog="open-llama",
        description="A TUI-based Ollama chat client.",
    )
    parser.add_argument(
        "--host",
        default=DEFAULT_BASE_URL,
        metavar="URL",
        help=f"Ollama server URL (default: {DEFAULT_BASE_URL})",
    )
    args = parser.parse_args()

    app = OpenLlamaApp(base_url=args.host)
    app.run()


if __name__ == "__main__":
    main()
