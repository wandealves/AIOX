import json
import time
import logging

from openai import AsyncOpenAI

from .base import LLMProvider, LLMResponse, ToolCallResult

logger = logging.getLogger(__name__)

MAX_TOOL_ITERATIONS = 10


class OpenAIProvider(LLMProvider):
    """OpenAI chat completions provider with tool/function calling support."""

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
        tools: list[dict] | None = None,
        tool_executor=None,
        attachments: list[dict] | None = None,
    ) -> LLMResponse:
        if not model:
            model = "gpt-4o-mini"

        if messages is None:
            messages = [
                {"role": "system", "content": system_prompt},
                {"role": "user", "content": user_message},
            ]

        # Append attachments to the last user message as multimodal content
        if attachments and messages:
            last_user_idx = None
            for i in range(len(messages) - 1, -1, -1):
                if messages[i]["role"] == "user":
                    last_user_idx = i
                    break
            if last_user_idx is not None:
                msg = messages[last_user_idx]
                content = msg["content"]
                # Convert to content array format
                if isinstance(content, str):
                    parts = [{"type": "text", "text": content}]
                else:
                    parts = list(content)
                for att in attachments:
                    if att["type"] == "image":
                        parts.append({
                            "type": "image_url",
                            "image_url": {"url": f"data:{att['content_type']};base64,{att['data_b64']}"},
                        })
                    else:
                        parts.append({
                            "type": "text",
                            "text": f"[Attached file: {att['filename']}]\n{att['content']}",
                        })
                messages[last_user_idx] = {"role": "user", "content": parts}

        start = time.monotonic()
        all_tools_called = []
        total_tokens = 0

        try:
            for iteration in range(MAX_TOOL_ITERATIONS):
                kwargs = {
                    "model": model,
                    "messages": messages,
                    "temperature": temperature,
                    "max_tokens": max_tokens,
                }
                if tools:
                    kwargs["tools"] = tools

                response = await self.client.chat.completions.create(**kwargs)

                if response.usage:
                    total_tokens += response.usage.total_tokens

                choice = response.choices[0]

                # Check if the model wants to call tools
                if choice.finish_reason == "tool_calls" and choice.message.tool_calls and tool_executor:
                    # Add assistant message with tool calls
                    messages.append(choice.message.model_dump())

                    for tool_call in choice.message.tool_calls:
                        fn = tool_call.function
                        try:
                            args = json.loads(fn.arguments) if fn.arguments else {}
                        except json.JSONDecodeError:
                            args = {}

                        # Find the tool definition to get endpoint info
                        tool_def = _find_tool(tools, fn.name)
                        if tool_def is None:
                            result = {"error": f"Tool '{fn.name}' not found"}
                            tool_result = ToolCallResult(
                                tool_name=fn.name,
                                arguments_json=fn.arguments or "{}",
                                result_json=json.dumps(result),
                                duration_ms=0,
                                error=result["error"],
                            )
                        else:
                            # Execute the tool
                            exec_result = await tool_executor.execute(
                                endpoint_url=tool_def.get("_endpoint_url", ""),
                                http_method=tool_def.get("_http_method", "POST"),
                                arguments=args,
                                headers=tool_def.get("_headers"),
                                auth_type=tool_def.get("_auth_type", ""),
                                auth_config=tool_def.get("_auth_config"),
                                timeout_sec=tool_def.get("_timeout_sec", 30),
                            )
                            tool_result = ToolCallResult(
                                tool_name=fn.name,
                                arguments_json=fn.arguments or "{}",
                                result_json=exec_result.get("result", ""),
                                duration_ms=exec_result.get("duration_ms", 0),
                                error=exec_result.get("error", ""),
                            )

                        all_tools_called.append(tool_result)

                        # Add tool result message
                        messages.append({
                            "role": "tool",
                            "tool_call_id": tool_call.id,
                            "content": tool_result.result_json or tool_result.error,
                        })

                    continue  # Re-call LLM with tool results

                # No tool calls — return final response
                text = choice.message.content or ""
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
            logger.error("OpenAI error: %s", e)
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
        fn = tool.get("function", {})
        if fn.get("name") == name:
            return tool
    return None
