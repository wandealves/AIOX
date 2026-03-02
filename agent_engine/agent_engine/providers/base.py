import json
from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from typing import Protocol, runtime_checkable

MAX_TOOL_ITERATIONS = 10


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


@runtime_checkable
class ToolExecutor(Protocol):
    """Protocol for tool executors passed to LLM providers."""

    async def execute(
        self,
        endpoint_url: str,
        http_method: str,
        arguments: dict,
        headers: dict | None,
        auth_type: str,
        auth_config: dict | None,
        timeout_sec: int,
        tool_type: str = "http",
        tool_name: str = "",
    ) -> dict: ...


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
        tool_executor: ToolExecutor | None = None,
        attachments: list[dict] | None = None,
    ) -> LLMResponse:
        """Generate a response from the LLM.

        If `messages` is provided, use the full messages array (with conversation
        history and memory context) instead of just system_prompt + user_message.

        If `tools` is provided (OpenAI/Anthropic format), enable function calling.
        tool_executor should implement the ToolExecutor protocol for running tool calls.
        """
        ...


def _find_tool(tools: list[dict], name: str) -> dict | None:
    """Find a tool definition by name in the tools list."""
    for tool in tools:
        fn = tool.get("function", {})
        if fn.get("name") == name:
            return tool
    return None


def _append_attachments(messages: list[dict], attachments: list[dict]) -> None:
    """Append attachments to the last user message as multimodal content parts.

    Replaces the last user message in-place with a multimodal content array.
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
                "type": "image_url",
                "image_url": {"url": f"data:{att['content_type']};base64,{att['data_b64']}"},
            })
        else:
            parts.append({
                "type": "text",
                "text": f"[Attached file: {att['filename']}]\n{att['content']}",
            })
    messages[last_user_idx] = {"role": "user", "content": parts}


def _apply_parameter_defaults(tool_def: dict, arguments: dict) -> dict:
    """Fill in missing arguments with default values from the tool's parameter schema."""
    params = tool_def.get("function", {}).get("parameters", {})
    properties = params.get("properties", {})
    merged = dict(arguments)
    for key, schema in properties.items():
        if key not in merged and "default" in schema:
            merged[key] = schema["default"]
    return merged


async def _execute_tool(
    tool_executor: ToolExecutor,
    tool_def: dict,
    tool_name: str,
    arguments: dict,
    args_json: str,
) -> ToolCallResult:
    """Execute a single tool call via the tool executor and return a ToolCallResult."""
    arguments = _apply_parameter_defaults(tool_def, arguments)
    exec_result = await tool_executor.execute(
        endpoint_url=tool_def.get("_endpoint_url", ""),
        http_method=tool_def.get("_http_method", "POST"),
        arguments=arguments,
        headers=tool_def.get("_headers"),
        auth_type=tool_def.get("_auth_type", ""),
        auth_config=tool_def.get("_auth_config"),
        timeout_sec=tool_def.get("_timeout_sec", 30),
        tool_type=tool_def.get("_tool_type", "http"),
        tool_name=tool_def.get("_tool_name", tool_name),
    )
    return ToolCallResult(
        tool_name=tool_name,
        arguments_json=args_json,
        result_json=exec_result.get("result", ""),
        duration_ms=exec_result.get("duration_ms", 0),
        error=exec_result.get("error", ""),
    )


def _make_tool_not_found_result(tool_name: str, args_json: str) -> ToolCallResult:
    """Build a ToolCallResult for a tool that was not found in the definitions list."""
    error_msg = f"Tool '{tool_name}' not found"
    return ToolCallResult(
        tool_name=tool_name,
        arguments_json=args_json,
        result_json=json.dumps({"error": error_msg}),
        duration_ms=0,
        error=error_msg,
    )
