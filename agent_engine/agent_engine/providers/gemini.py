import json
import logging
import time

from .base import (
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

DEFAULT_MODEL = "gemini-2.0-flash"


class GeminiProvider(LLMProvider):
    """Google Gemini provider using the google-genai SDK.

    Receives tools in OpenAI format (with private _endpoint_url metadata) and
    converts to Gemini FunctionDeclaration format internally before each API call.
    Roles are mapped: 'assistant' → 'model', 'system' → system_instruction config.
    """

    def __init__(self, api_key: str):
        from google import genai  # lazy import to avoid hard dependency at module load

        self._client = genai.Client(api_key=api_key)

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

        # Gemini uses system_instruction as a separate config field.
        # Convert OpenAI-style roles to Gemini's user/model roles.
        system_instruction = system_prompt
        if messages is not None:
            chat_contents: list[dict] = []
            for msg in list(messages):
                if msg["role"] == "system":
                    system_instruction = _extract_text_content(msg["content"])
                    continue
                role = "model" if msg["role"] == "assistant" else "user"
                chat_contents.append({
                    "role": role,
                    "parts": _to_gemini_parts(msg["content"]),
                })
        else:
            chat_contents = [{"role": "user", "parts": [{"text": user_message}]}]

        if attachments:
            _append_gemini_attachments(chat_contents, attachments)

        # Convert from OpenAI tool format to Gemini Tool format.
        # The original `tools` list is kept for execution metadata lookups.
        api_tools = _to_gemini_api_tools(tools) if tools else None

        start = time.monotonic()
        all_tools_called: list[ToolCallResult] = []
        total_tokens = 0

        try:
            from google.genai import types

            for _ in range(MAX_TOOL_ITERATIONS):
                config = types.GenerateContentConfig(
                    system_instruction=system_instruction,
                    temperature=temperature,
                    max_output_tokens=max_tokens,
                    tools=api_tools if api_tools and tool_executor else None,
                )

                response = await self._client.aio.models.generate_content(
                    model=model,
                    contents=chat_contents,
                    config=config,
                )

                if response.usage_metadata:
                    total_tokens += (
                        (response.usage_metadata.prompt_token_count or 0)
                        + (response.usage_metadata.candidates_token_count or 0)
                    )

                candidate = response.candidates[0] if response.candidates else None
                function_calls = []
                if candidate and candidate.content and candidate.content.parts:
                    function_calls = [
                        p.function_call
                        for p in candidate.content.parts
                        if p.function_call
                    ]

                if function_calls and tool_executor:
                    # Add the model's response (may contain text + function_call parts)
                    chat_contents.append({
                        "role": "model",
                        "parts": _parts_to_dict(candidate.content.parts),
                    })

                    tool_response_parts = []
                    for fc in function_calls:
                        args = dict(fc.args) if fc.args else {}
                        args_json = json.dumps(args)

                        tool_def = _find_tool(tools, fc.name)
                        if tool_def is None:
                            tool_result = _make_tool_not_found_result(fc.name, args_json)
                            result_response: dict = {"error": tool_result.error}
                        else:
                            tool_result = await _execute_tool(
                                tool_executor, tool_def, fc.name, args, args_json
                            )
                            try:
                                result_response = (
                                    json.loads(tool_result.result_json)
                                    if tool_result.result_json
                                    else {}
                                )
                            except json.JSONDecodeError:
                                result_response = {"result": tool_result.result_json}

                        all_tools_called.append(tool_result)
                        tool_response_parts.append({
                            "function_response": {
                                "name": fc.name,
                                "response": result_response,
                            }
                        })

                    chat_contents.append({"role": "user", "parts": tool_response_parts})
                    continue

                # Extract text (ignoring function_call parts if any remain)
                text = ""
                if candidate and candidate.content and candidate.content.parts:
                    text = "".join(
                        p.text for p in candidate.content.parts
                        if hasattr(p, "text") and p.text and not p.function_call
                    )

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
            logger.error("Gemini error: %s", e)
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


def _to_gemini_parts(content) -> list[dict]:
    """Convert a message content value to Gemini parts format.

    Handles plain strings and OpenAI-style multimodal content arrays
    (text + image_url with data URIs).
    """
    if isinstance(content, str):
        return [{"text": content}]

    if isinstance(content, list):
        parts: list[dict] = []
        for item in content:
            if not isinstance(item, dict):
                continue
            kind = item.get("type", "")
            if kind == "text":
                parts.append({"text": item["text"]})
            elif kind == "image_url":
                url = item.get("image_url", {}).get("url", "")
                if url.startswith("data:"):
                    # data:<mime>;base64,<data>
                    header, b64 = url.split(",", 1)
                    mime = header.split(";")[0].replace("data:", "")
                    parts.append({"inline_data": {"mime_type": mime, "data": b64}})
        return parts or [{"text": ""}]

    return [{"text": str(content)}]


def _parts_to_dict(parts) -> list[dict]:
    """Convert Gemini SDK Part objects to plain dicts for inclusion in chat_contents."""
    result: list[dict] = []
    for part in parts:
        if part.function_call:
            result.append({
                "function_call": {
                    "name": part.function_call.name,
                    "args": dict(part.function_call.args) if part.function_call.args else {},
                }
            })
        elif hasattr(part, "text") and part.text:
            result.append({"text": part.text})
    return result


def _append_gemini_attachments(contents: list[dict], attachments: list[dict]) -> None:
    """Append attachments to the last user message in Gemini inline_data format."""
    last_user_idx = None
    for i in range(len(contents) - 1, -1, -1):
        if contents[i]["role"] == "user":
            last_user_idx = i
            break
    if last_user_idx is None:
        return

    parts = list(contents[last_user_idx]["parts"])
    for att in attachments:
        if att["type"] == "image":
            parts.append({
                "inline_data": {
                    "mime_type": att["content_type"],
                    "data": att["data_b64"],
                }
            })
        else:
            parts.append({
                "text": f"[Attached file: {att['filename']}]\n{att['content']}"
            })
    contents[last_user_idx] = {"role": "user", "parts": parts}


def _to_gemini_api_tools(tools: list[dict]):
    """Convert OpenAI-formatted tool definitions to a Gemini Tool object.

    Private metadata fields (_endpoint_url, etc.) are not included — they stay
    in the original `tools` list for execution via _find_tool + _execute_tool.
    """
    from google.genai import types

    function_declarations = []
    for tool in tools:
        fn = tool.get("function", {})
        function_declarations.append(
            types.FunctionDeclaration(
                name=fn.get("name", ""),
                description=fn.get("description", ""),
                parameters=fn.get("parameters", {"type": "object", "properties": {}}),
            )
        )
    return types.Tool(function_declarations=function_declarations)
