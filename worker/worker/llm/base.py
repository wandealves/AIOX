from abc import ABC, abstractmethod
from dataclasses import dataclass, field


@dataclass
class ToolCallResult:
    """Result of a tool call during LLM execution."""
    tool_name: str
    arguments_json: str
    result_json: str
    duration_ms: int
    error: str = ""


@dataclass
class LLMResponse:
    """Response from an LLM provider."""
    text: str
    tokens_used: int
    model_used: str
    duration_ms: int
    error: str = ""
    tools_called: list[ToolCallResult] = field(default_factory=list)


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
        tools: list[dict] | None = None,
        tool_executor=None,
        attachments: list[dict] | None = None,
    ) -> LLMResponse:
        """Generate a response from the LLM.

        If `messages` is provided, use the full messages array (with conversation
        history and memory context) instead of just system_prompt + user_message.

        If `tools` is provided (OpenAI/Anthropic format), enable function calling.
        tool_executor should have an `execute()` method for running tool calls.
        """
        ...
