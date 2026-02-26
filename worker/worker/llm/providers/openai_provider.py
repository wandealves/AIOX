import json
import logging
import time

from openai import AsyncOpenAI

from .llm_provider import (
    LLMProvider,
    LLMResponse,
    ToolCallResult,
    ToolExecutor,
    MAX_TOOL_ITERATIONS,
    _find_tool,
    _append_attachments,
    _execute_tool,
    _make_tool_not_found_result,
)

logger = logging.getLogger(__name__)

DEFAULT_MODEL = "gpt-4o-mini"


class OpenAIProvider(LLMProvider):
    """OpenAI chat completions provider with tool/function calling support."""

    def __init__(self, api_key: str, base_url: str | None = None):
        """Initialize OpenAI provider.

        Args:
            api_key: OpenAI API key.
            base_url: Optional base URL for compatible APIs (e.g., Azure, local proxies).
        """
        self.base_url = base_url
        client_kwargs: dict = {"api_key": api_key}
        if base_url:
            client_kwargs["base_url"] = base_url
        self.client = AsyncOpenAI(**client_kwargs)

    async def generate(
        self,
        system_prompt: str,
        user_message: str,
        model: str = "",
        temperature: float = 0.7,
        max_tokens: int = 1024,
        messages: list[dict] | None = None,
        tools: list[dict] | None = None,
        tool_executor: ToolExecutor | None = None,
        attachments: list[dict] | None = None,
    ) -> LLMResponse:
        model = model or DEFAULT_MODEL
        # Work on a copy to avoid mutating the caller's list
        messages = list(messages) if messages is not None else [
            {"role": "system", "content": system_prompt},
            {"role": "user", "content": user_message},
        ]

        if attachments:
            _append_attachments(messages, attachments)

        start = time.monotonic()
        all_tools_called: list[ToolCallResult] = []
        total_tokens = 0

        try:
            for _ in range(MAX_TOOL_ITERATIONS):
                request_kwargs: dict = {
                    "model": model,
                    "messages": messages,
                    "temperature": temperature,
                    "max_tokens": max_tokens,
                }
                if tools:
                    request_kwargs["tools"] = tools

                response = await self.client.chat.completions.create(**request_kwargs)

                if response.usage:
                    total_tokens += response.usage.total_tokens

                choice = response.choices[0]

                if choice.finish_reason == "tool_calls" and choice.message.tool_calls and tool_executor:
                    messages.append(choice.message.model_dump())

                    for tool_call in choice.message.tool_calls:
                        fn = tool_call.function
                        try:
                            args = json.loads(fn.arguments) if fn.arguments else {}
                        except json.JSONDecodeError:
                            args = {}

                        tool_def = _find_tool(tools, fn.name)
                        if tool_def is None:
                            tool_result = _make_tool_not_found_result(fn.name, fn.arguments or "{}")
                        else:
                            tool_result = await _execute_tool(
                                tool_executor, tool_def, fn.name, args, fn.arguments or "{}"
                            )

                        all_tools_called.append(tool_result)
                        messages.append({
                            "role": "tool",
                            "tool_call_id": tool_call.id,
                            "content": tool_result.result_json or tool_result.error,
                        })

                    continue

                duration_ms = int((time.monotonic() - start) * 1000)
                return LLMResponse(
                    text=choice.message.content or "",
                    tokens_used=total_tokens,
                    model_used=model,
                    duration_ms=duration_ms,
                    tools_called=all_tools_called,
                )

            # Max iterations reached
            duration_ms = int((time.monotonic() - start) * 1000)
            return LLMResponse(
                text="Tool calling loop exceeded maximum iterations.",
                tokens_used=total_tokens,
                model_used=model,
                duration_ms=duration_ms,
                error="max tool iterations exceeded",
                tools_called=all_tools_called,
            )

        except Exception as e:
            duration_ms = int((time.monotonic() - start) * 1000)
            logger.error("OpenAI error: %s", e)
            return LLMResponse(
                text="",
                tokens_used=total_tokens,
                model_used=model,
                duration_ms=duration_ms,
                error=str(e),
                tools_called=all_tools_called,
            )
