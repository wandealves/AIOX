import json
import logging
import time

import httpx

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


class OllamaProvider(LLMProvider):
    """Ollama chat completions provider with tool/function calling support."""

    def __init__(self, model: str, base_url: str = "http://localhost:11434"):
        self.model = model
        self.base_url = base_url
        # Reuse a single client across calls for connection pooling
        self._client = httpx.AsyncClient()

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
        model = model or self.model
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
                payload: dict = {
                    "model": model,
                    "messages": messages,
                    "options": {
                        "temperature": temperature,
                        "num_predict": max_tokens,
                    },
                    "stream": False,
                }
                if tools and tool_executor:
                    payload["tools"] = tools

                response = await self._client.post(
                    f"{self.base_url}/api/chat",
                    json=payload,
                    timeout=120.0,
                )
                response.raise_for_status()
                data = response.json()

                total_tokens += data.get("prompt_eval_count", 0) + data.get("eval_count", 0)

                message = data.get("message", {})
                assistant_content = message.get("content", "")
                tool_calls = message.get("tool_calls", [])

                if tool_calls and tool_executor:
                    messages.append({
                        "role": "assistant",
                        "content": assistant_content,
                        "tool_calls": tool_calls,
                    })

                    for tool_call in tool_calls:
                        fn = tool_call.get("function", {})
                        tool_name = fn.get("name", "")
                        arguments = fn.get("arguments", {})
                        args_json = (
                            json.dumps(arguments) if isinstance(arguments, dict) else str(arguments)
                        )

                        tool_def = _find_tool(tools, tool_name)
                        if tool_def is None:
                            tool_result = _make_tool_not_found_result(tool_name, args_json)
                        else:
                            arguments_dict = arguments if isinstance(arguments, dict) else {}
                            tool_result = await _execute_tool(
                                tool_executor, tool_def, tool_name, arguments_dict, args_json
                            )

                        all_tools_called.append(tool_result)
                        messages.append({
                            "role": "tool",
                            "name": tool_name,
                            "content": tool_result.result_json or tool_result.error,
                            "tool_call_id": tool_call.get("id", ""),
                        })

                    continue

                duration_ms = int((time.monotonic() - start) * 1000)
                return LLMResponse(
                    text=assistant_content,
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

        except httpx.HTTPStatusError as e:
            duration_ms = int((time.monotonic() - start) * 1000)
            logger.error("Ollama HTTP error: %s", e)
            return LLMResponse(
                text="",
                tokens_used=total_tokens,
                model_used=model,
                duration_ms=duration_ms,
                error=f"HTTP {e.response.status_code}: {e.response.text}",
                tools_called=all_tools_called,
            )
        except httpx.RequestError as e:
            duration_ms = int((time.monotonic() - start) * 1000)
            logger.error("Ollama request error: %s", e)
            return LLMResponse(
                text="",
                tokens_used=total_tokens,
                model_used=model,
                duration_ms=duration_ms,
                error=str(e),
                tools_called=all_tools_called,
            )
        except Exception as e:
            duration_ms = int((time.monotonic() - start) * 1000)
            logger.error("Ollama error: %s", e)
            return LLMResponse(
                text="",
                tokens_used=total_tokens,
                model_used=model,
                duration_ms=duration_ms,
                error=str(e),
                tools_called=all_tools_called,
            )
