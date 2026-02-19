import time
import logging

from openai import AsyncOpenAI

from .base import LLMProvider, LLMResponse

logger = logging.getLogger(__name__)


class OpenAIProvider(LLMProvider):
    """OpenAI chat completions provider."""

    def __init__(self, api_key: str):
        self.client = AsyncOpenAI(api_key=api_key)

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
            model = "gpt-4o-mini"

        # Use full messages array if provided, otherwise build simple two-message array
        if messages is None:
            messages = [
                {"role": "system", "content": system_prompt},
                {"role": "user", "content": user_message},
            ]

        start = time.monotonic()
        try:
            response = await self.client.chat.completions.create(
                model=model,
                messages=messages,
                temperature=temperature,
                max_tokens=max_tokens,
            )
            duration_ms = int((time.monotonic() - start) * 1000)

            tokens = 0
            if response.usage:
                tokens = response.usage.total_tokens

            text = response.choices[0].message.content or ""

            return LLMResponse(
                text=text,
                tokens_used=tokens,
                model_used=model,
                duration_ms=duration_ms,
            )
        except Exception as e:
            duration_ms = int((time.monotonic() - start) * 1000)
            logger.error("OpenAI error: %s", e)
            return LLMResponse(
                text="",
                tokens_used=0,
                model_used=model,
                duration_ms=duration_ms,
                error=str(e),
            )
