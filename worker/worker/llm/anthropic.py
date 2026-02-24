import json
import time
import logging

from anthropic import AsyncAnthropic

from .base import LLMProvider, LLMResponse, ToolCallResult

logger = logging.getLogger(__name__)

MAX_TOOL_ITERATIONS = 10


class AnthropicProvider(LLMProvider):
    """Anthropic messages API provider with tool/function calling support."""

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
        tools: list[dict] | None = None,
        tool_executor=None,
        attachments: list[dict] | None = None,
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

        # Append attachments to the last user message as multimodal content
        if attachments and chat_messages:
            last_user_idx = None
            for i in range(len(chat_messages) - 1, -1, -1):
                if chat_messages[i]["role"] == "user":
                    last_user_idx = i
                    break
            if last_user_idx is not None:
                msg = chat_messages[last_user_idx]
                content = msg["content"]
                if isinstance(content, str):
                    parts = [{"type": "text", "text": content}]
                else:
                    parts = list(content)
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
                chat_messages[last_user_idx] = {"role": "user", "content": parts}

        start = time.monotonic()
        all_tools_called = []
        total_tokens = 0

        try:
            for iteration in range(MAX_TOOL_ITERATIONS):
                kwargs = {
                    "model": model,
                    "max_tokens": max_tokens,
                    "system": system,
                    "messages": chat_messages,
                    "temperature": temperature,
                }
                if tools:
                    kwargs["tools"] = tools

                response = await self.client.messages.create(**kwargs)

                if response.usage:
                    total_tokens += response.usage.input_tokens + response.usage.output_tokens

                # Check for tool_use blocks
                tool_use_blocks = [b for b in response.content if b.type == "tool_use"]
                text_blocks = [b for b in response.content if b.type == "text"]

                if tool_use_blocks and tool_executor and response.stop_reason == "tool_use":
                    # Add assistant message
                    chat_messages.append({
                        "role": "assistant",
                        "content": [b.model_dump() for b in response.content],
                    })

                    tool_results = []
                    for tool_block in tool_use_blocks:
                        args = tool_block.input or {}
                        args_json = json.dumps(args) if isinstance(args, dict) else str(args)

                        # Find tool definition
                        tool_def = _find_tool(tools, tool_block.name)
                        if tool_def is None:
                            result = {"error": f"Tool '{tool_block.name}' not found"}
                            tool_result = ToolCallResult(
                                tool_name=tool_block.name,
                                arguments_json=args_json,
                                result_json=json.dumps(result),
                                duration_ms=0,
                                error=result["error"],
                            )
                        else:
                            exec_result = await tool_executor.execute(
                                endpoint_url=tool_def.get("_endpoint_url", ""),
                                http_method=tool_def.get("_http_method", "POST"),
                                arguments=args if isinstance(args, dict) else {},
                                headers=tool_def.get("_headers"),
                                auth_type=tool_def.get("_auth_type", ""),
                                auth_config=tool_def.get("_auth_config"),
                                timeout_sec=tool_def.get("_timeout_sec", 30),
                            )
                            tool_result = ToolCallResult(
                                tool_name=tool_block.name,
                                arguments_json=args_json,
                                result_json=exec_result.get("result", ""),
                                duration_ms=exec_result.get("duration_ms", 0),
                                error=exec_result.get("error", ""),
                            )

                        all_tools_called.append(tool_result)
                        tool_results.append({
                            "type": "tool_result",
                            "tool_use_id": tool_block.id,
                            "content": tool_result.result_json or tool_result.error,
                        })

                    # Add tool results as user message
                    chat_messages.append({"role": "user", "content": tool_results})
                    continue  # Re-call LLM

                # No tool calls — extract text
                text = ""
                for block in text_blocks:
                    text += block.text

                duration_ms = int((time.monotonic() - start) * 1000)
                return LLMResponse(
                    text=text,
                    tokens_used=total_tokens,
                    model_used=model,
                    duration_ms=duration_ms,
                    tools_called=all_tools_called,
                )

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


def _find_tool(tools: list[dict], name: str) -> dict | None:
    """Find a tool definition by name."""
    for tool in tools:
        if tool.get("name") == name:
            return tool
    return None
