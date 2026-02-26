import json
import logging
import time

from anthropic import AsyncAnthropic

from .llm_provider import (
    LLMProvider,
    LLMResponse,
    ToolCallResult,
    ToolExecutor,
    MAX_TOOL_ITERATIONS,
    _find_tool,
    _execute_tool,
    _make_tool_not_found_result,
)

logger = logging.getLogger(__name__)

DEFAULT_MODEL = "claude-sonnet-4-20250514"


class AnthropicProvider(LLMProvider):
    """Anthropic messages API provider with tool/function calling support.

    Receives tools in OpenAI format (with private _endpoint_url metadata) and
    converts to Anthropic format internally before each API call.
    """

    def __init__(self, api_key: str):
        self._client = AsyncAnthropic(api_key=api_key)

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

        # Anthropic requires system as a top-level parameter, not a message role.
        # Separate system content from chat messages.
        system = system_prompt
        if messages is not None:
            chat_messages = []
            for msg in list(messages):
                if msg["role"] == "system":
                    system = _extract_text_content(msg["content"])
                else:
                    chat_messages.append(msg)
        else:
            chat_messages = [{"role": "user", "content": user_message}]

        if attachments:
            _append_anthropic_attachments(chat_messages, attachments)

        # Convert from OpenAI tool format to Anthropic API format (for the API call only).
        # The original `tools` list is kept for execution metadata lookups.
        api_tools = _to_anthropic_api_tools(tools) if tools else None

        start = time.monotonic()
        all_tools_called: list[ToolCallResult] = []
        total_tokens = 0

        try:
            for _ in range(MAX_TOOL_ITERATIONS):
                request_kwargs: dict = {
                    "model": model,
                    "max_tokens": max_tokens,
                    "system": system,
                    "messages": chat_messages,
                    "temperature": temperature,
                }
                if api_tools and tool_executor:
                    request_kwargs["tools"] = api_tools

                response = await self._client.messages.create(**request_kwargs)

                if response.usage:
                    total_tokens += response.usage.input_tokens + response.usage.output_tokens

                tool_use_blocks = [b for b in response.content if b.type == "tool_use"]

                if tool_use_blocks and tool_executor and response.stop_reason == "tool_use":
                    # Add the assistant message (with tool_use blocks) to history
                    chat_messages.append({
                        "role": "assistant",
                        "content": [b.model_dump() for b in response.content],
                    })

                    tool_results = []
                    for block in tool_use_blocks:
                        args = dict(block.input) if isinstance(block.input, dict) else {}
                        args_json = json.dumps(args)

                        tool_def = _find_tool(tools, block.name)
                        if tool_def is None:
                            tool_result = _make_tool_not_found_result(block.name, args_json)
                        else:
                            tool_result = await _execute_tool(
                                tool_executor, tool_def, block.name, args, args_json
                            )

                        all_tools_called.append(tool_result)
                        tool_results.append({
                            "type": "tool_result",
                            "tool_use_id": block.id,
                            "content": tool_result.result_json or tool_result.error,
                        })

                    chat_messages.append({"role": "user", "content": tool_results})
                    continue

                text = "".join(b.text for b in response.content if b.type == "text")
                duration_ms = int((time.monotonic() - start) * 1000)
                return LLMResponse(
                    text=text,
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
            logger.error("Anthropic error: %s", e)
            return LLMResponse(
                text="",
                tokens_used=total_tokens,
                model_used=model,
                duration_ms=duration_ms,
                error=str(e),
                tools_called=all_tools_called,
            )


# ---------------------------------------------------------------------------
# Private helpers
# ---------------------------------------------------------------------------

def _extract_text_content(content) -> str:
    """Extract plain text from a message content that may be a string or parts list."""
    if isinstance(content, str):
        return content
    if isinstance(content, list):
        return " ".join(
            p.get("text", "") for p in content
            if isinstance(p, dict) and p.get("type") == "text"
        )
    return str(content)


def _append_anthropic_attachments(messages: list[dict], attachments: list[dict]) -> None:
    """Append attachments to the last user message in Anthropic multimodal format.

    Images use the base64 source format; other files are appended as text parts.
    """
    last_user_idx = None
    for i in range(len(messages) - 1, -1, -1):
        if messages[i]["role"] == "user":
            last_user_idx = i
            break
    if last_user_idx is None:
        return

    content = messages[last_user_idx]["content"]
    parts: list[dict] = (
        [{"type": "text", "text": content}] if isinstance(content, str) else list(content)
    )
    for att in attachments:
        if att["type"] == "image":
            parts.append({
                "type": "image",
                "source": {
                    "type": "base64",
                    "media_type": att["content_type"],
                    "data": att["data_b64"],
                },
            })
        else:
            parts.append({
                "type": "text",
                "text": f"[Attached file: {att['filename']}]\n{att['content']}",
            })
    messages[last_user_idx] = {"role": "user", "content": parts}


def _to_anthropic_api_tools(tools: list[dict]) -> list[dict]:
    """Convert OpenAI-formatted tool definitions to Anthropic API format.

    Private metadata fields (_endpoint_url, etc.) are not included — they stay
    in the original `tools` list for execution via _find_tool + _execute_tool.
    """
    result = []
    for tool in tools:
        fn = tool.get("function", {})
        result.append({
            "name": fn.get("name", ""),
            "description": fn.get("description", ""),
            "input_schema": fn.get("parameters", {"type": "object", "properties": {}}),
        })
    return result
