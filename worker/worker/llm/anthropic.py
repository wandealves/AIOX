import time
import logging

from anthropic import AsyncAnthropic

from .base import LLMProvider, LLMResponse

logger = logging.getLogger(__name__)


class AnthropicProvider(LLMProvider):
    """Anthropic messages API provider."""

    def __init__(self, api_key: str):
        self.client = AsyncAnthropic(api_key=api_key)

    async def generate(
        self,
        system_prompt: str,
        user_message: str,
        model: str = "",
        temperature: float = 0.7,
        max_tokens: int = 1024,
        messages: list[dict] | None = None,
    ) -> LLMResponse:
        if not model:
            model = "claude-sonnet-4-20250514"

        # For Anthropic, extract system from messages[0] if full messages provided
        system = system_prompt
        chat_messages = [{"role": "user", "content": user_message}]
        if messages is not None:
            system = ""
            chat_messages = []
            for msg in messages:
                if msg["role"] == "system":
                    system = msg["content"]
                else:
                    chat_messages.append(msg)

        start = time.monotonic()
        try:
            response = await self.client.messages.create(
                model=model,
                max_tokens=max_tokens,
                system=system,
                messages=chat_messages,
                temperature=temperature,
            )
            duration_ms = int((time.monotonic() - start) * 1000)

            tokens = 0
            if response.usage:
                tokens = response.usage.input_tokens + response.usage.output_tokens

            text = ""
            for block in response.content:
                if block.type == "text":
                    text += block.text

            return LLMResponse(
                text=text,
                tokens_used=tokens,
                model_used=model,
                duration_ms=duration_ms,
            )
        except Exception as e:
            duration_ms = int((time.monotonic() - start) * 1000)
            logger.error("Anthropic error: %s", e)
            return LLMResponse(
                text="",
                tokens_used=0,
                model_used=model,
                duration_ms=duration_ms,
                error=str(e),
            )
