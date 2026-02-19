from abc import ABC, abstractmethod
from dataclasses import dataclass


@dataclass
class LLMResponse:
    """Response from an LLM provider."""
    text: str
    tokens_used: int
    model_used: str
    duration_ms: int
    error: str = ""


class LLMProvider(ABC):
    """Abstract base class for LLM providers."""

    @abstractmethod
    async def generate(
        self,
        system_prompt: str,
        user_message: str,
        model: str = "",
        temperature: float = 0.7,
        max_tokens: int = 1024,
        messages: list[dict] | None = None,
    ) -> LLMResponse:
        """Generate a response from the LLM.

        If `messages` is provided, use the full messages array (with conversation
        history and memory context) instead of just system_prompt + user_message.
        """
        ...
