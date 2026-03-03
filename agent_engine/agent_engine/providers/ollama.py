import json
import logging
import time

import httpx

from ..errors import (
    ProviderOverloadedError,
    ProviderTimeoutError,
)
from .base import (
    MAX_TOOL_ITERATIONS,
    LLMProvider,
    LLMResponse,
    ToolCallResult,
    ToolExecutor,
    _append_attachments,
    _execute_tool,
    _find_tool,
    _make_tool_not_found_result,
)

logger = logging.getLogger(__name__)

# Parameter names that carry credentials and must NOT be shown to the LLM.
_SENSITIVE_PARAMS = frozenset(
    {
        "api_key",
        "apikey",
        "api_secret",
        "secret",
        "token",
        "access_token",
        "refresh_token",
        "password",
        "credentials",
    }
)


def _strip_tools_for_ollama(tools: list[dict]) -> list[dict]:
    """Build a clean tools list for the Ollama API.

    - Removes private metadata fields (_endpoint_url, _auth_config, …)
    - Removes sensitive parameters (api_key, token, …) from the schema
      so the LLM doesn't hallucinate placeholder values for credentials.
    """
    cleaned = []
    for tool in tools:
        fn = tool.get("function", {})
        params = fn.get("parameters", {})
        properties = {
            k: v
            for k, v in params.get("properties", {}).items()
            if k.lower() not in _SENSITIVE_PARAMS
        }
        required = [
            r for r in params.get("required", []) if r.lower() not in _SENSITIVE_PARAMS
        ]
        clean_params = {**params, "properties": properties, "required": required}

        cleaned.append(
            {
                "type": "function",
                "function": {
                    "name": fn.get("name", ""),
                    "description": fn.get("description", ""),
                    "parameters": clean_params,
                },
            }
        )
    return cleaned


def _parse_text_tool_calls(text: str, tools: list[dict] | None) -> list[dict]:
    """Try to extract tool calls from plain text when the model doesn't use structured tool_calls.

    Many Ollama models (especially smaller ones) output tool calls as JSON in their
    text response instead of using the structured tool_calls API field.

    Recognizes patterns like:
      {"name": "tool_name", "parameters": {...}}
      {"name": "tool_name", "arguments": {...}}
      [{"name": "tool_name", ...}]
    Also handles the model prefixing text before the JSON (e.g. "Ollama\\n{...}").
    """
    if not tools or not text:
        return []

    # Collect known tool names for validation
    known_names = set()
    for tool in tools:
        fn = tool.get("function", {})
        name = fn.get("name", "")
        if name:
            known_names.add(name)
            # Also add hyphenated variant
            known_names.add(name.replace("_", "-"))

    # Try to find JSON object(s) in the text
    results = []
    # Find all potential JSON objects
    brace_depth = 0
    json_start = None
    for i, ch in enumerate(text):
        if ch == "{" and brace_depth == 0:
            json_start = i
            brace_depth = 1
        elif ch == "{":
            brace_depth += 1
        elif ch == "}" and brace_depth > 0:
            brace_depth -= 1
            if brace_depth == 0 and json_start is not None:
                candidate = text[json_start : i + 1]
                parsed = _try_parse_tool_call(candidate, known_names)
                if parsed:
                    results.append(parsed)
                json_start = None

    return results


def _try_parse_tool_call(candidate: str, known_names: set[str]) -> dict | None:
    """Try to parse a JSON string as a tool call. Returns an Ollama-style tool_call dict or None."""
    try:
        data = json.loads(candidate)
    except json.JSONDecodeError:
        return None

    if not isinstance(data, dict):
        return None

    name = data.get("name", "")
    if not name:
        return None

    # Check if name matches a known tool (exact or normalized)
    normalized = name.replace("-", "_").lower()
    if name not in known_names and normalized not in {n.replace("-", "_").lower() for n in known_names}:
        return None

    # Accept "parameters", "arguments", or "input" as the args field
    args = data.get("parameters") or data.get("arguments") or data.get("input") or {}
    if not isinstance(args, dict):
        args = {}

    return {
        "function": {
            "name": normalized,
            "arguments": args,
        }
    }


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
        messages = (
            list(messages)
            if messages is not None
            else [
                {"role": "system", "content": system_prompt},
                {"role": "user", "content": user_message},
            ]
        )

        # Clean version for the LLM (no private metadata, no credentials).
        # The original `tools` list is kept for _find_tool / _execute_tool.
        llm_tools = _strip_tools_for_ollama(tools) if tools else None

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
                if llm_tools and tool_executor:
                    payload["tools"] = llm_tools

                response = await self._client.post(
                    f"{self.base_url}/api/chat",
                    json=payload,
                    timeout=120.0,
                )
                response.raise_for_status()
                data = response.json()

                total_tokens += data.get("eval_count", 0)

                message = data.get("message", {})
                assistant_content = message.get("content", "")
                tool_calls = message.get("tool_calls", [])

                # Fallback: if model emitted tool calls as plain text instead
                # of using the structured tool_calls field, try to parse them.
                if not tool_calls and tool_executor and tools and assistant_content:
                    text_tool_calls = _parse_text_tool_calls(assistant_content, llm_tools)
                    if text_tool_calls:
                        logger.info(
                            "Parsed %d text-based tool call(s) from Ollama response",
                            len(text_tool_calls),
                        )
                        tool_calls = text_tool_calls

                if tool_calls and tool_executor:
                    messages.append(
                        {
                            "role": "assistant",
                            "content": assistant_content,
                            "tool_calls": tool_calls,
                        }
                    )

                    for tool_call in tool_calls:
                        fn = tool_call.get("function", {})
                        tool_name = fn.get("name", "")
                        arguments = fn.get("arguments", {})
                        args_json = (
                            json.dumps(arguments)
                            if isinstance(arguments, dict)
                            else str(arguments)
                        )

                        tool_def = _find_tool(tools, tool_name)
                        if tool_def is None:
                            tool_result = _make_tool_not_found_result(
                                tool_name, args_json
                            )
                        else:
                            arguments_dict = (
                                arguments if isinstance(arguments, dict) else {}
                            )
                            tool_result = await _execute_tool(
                                tool_executor,
                                tool_def,
                                tool_name,
                                arguments_dict,
                                args_json,
                            )

                        all_tools_called.append(tool_result)
                        messages.append(
                            {
                                "role": "tool",
                                "name": tool_name,
                                "content": tool_result.result_json or tool_result.error,
                                "tool_call_id": tool_call.get("id", ""),
                            }
                        )

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
            status = e.response.status_code
            if status == 429:
                raise ProviderOverloadedError(f"Ollama rate limited: {e}") from e
            if status >= 500:
                raise ProviderOverloadedError(
                    f"Ollama server error ({status}): {e}"
                ) from e
            duration_ms = int((time.monotonic() - start) * 1000)
            logger.error("Ollama HTTP error: %s", e)
            return LLMResponse(
                text="",
                tokens_used=total_tokens,
                model_used=model,
                duration_ms=duration_ms,
                error=f"HTTP {status}: {e.response.text}",
                tools_called=all_tools_called,
            )
        except httpx.TimeoutException as e:
            raise ProviderTimeoutError(f"Ollama timeout: {e}") from e
        except httpx.RequestError as e:
            raise ProviderTimeoutError(f"Ollama connection error: {e}") from e
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
