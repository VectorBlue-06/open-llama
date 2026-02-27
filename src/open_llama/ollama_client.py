"""Async HTTP client for the Ollama REST API."""

from __future__ import annotations

import json
from typing import AsyncIterator

import httpx

DEFAULT_BASE_URL = "http://localhost:11434"


class OllamaClient:
    """Thin async wrapper around the Ollama REST API."""

    def __init__(self, base_url: str = DEFAULT_BASE_URL) -> None:
        self._base_url = base_url.rstrip("/")

    # ------------------------------------------------------------------
    # Model discovery
    # ------------------------------------------------------------------

    async def list_models(self) -> list[str]:
        """Return the names of all locally-available models."""
        async with httpx.AsyncClient() as client:
            response = await client.get(
                f"{self._base_url}/api/tags",
                timeout=10.0,
            )
            response.raise_for_status()
            data = response.json()
            return [m["name"] for m in data.get("models", [])]

    # ------------------------------------------------------------------
    # Streaming chat
    # ------------------------------------------------------------------

    async def chat_stream(
        self,
        model: str,
        messages: list[dict[str, str]],
    ) -> AsyncIterator[str]:
        """Yield text tokens from a streaming chat completion.

        Each yielded value is a partial response string.
        """
        payload = {
            "model": model,
            "messages": messages,
            "stream": True,
        }
        async with httpx.AsyncClient() as client:
            async with client.stream(
                "POST",
                f"{self._base_url}/api/chat",
                json=payload,
                timeout=None,
            ) as response:
                response.raise_for_status()
                async for line in response.aiter_lines():
                    if not line:
                        continue
                    chunk = json.loads(line)
                    content = chunk.get("message", {}).get("content", "")
                    if content:
                        yield content
                    if chunk.get("done"):
                        return
