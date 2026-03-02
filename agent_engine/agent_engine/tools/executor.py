"""Tool executor with HTTP and built-in tool support."""

import asyncio
import json
import logging
import time

import aiohttp

from . import registry

logger = logging.getLogger(__name__)


class ToolExecutor:
    """Executes tool calls for LLM function calling.

    Routes to built-in tools when a matching tool is registered
    in the registry, otherwise falls back to HTTP execution.
    """

    def __init__(self, tool_context: dict | None = None):
        self.tool_context = tool_context or {}

    async def execute(
        self,
        endpoint_url: str,
        http_method: str,
        arguments: dict,
        headers: dict | None = None,
        auth_type: str = "",
        auth_config: dict | None = None,
        timeout_sec: int = 30,
        tool_type: str = "http",
        tool_name: str = "",
    ) -> dict:
        """Execute a tool call and return the result.

        Returns dict with keys: result, duration_ms, error.
        """
        builtin = registry.get_tool(tool_name)
        if builtin:
            return await self._execute_builtin(
                builtin, arguments, auth_type=auth_type, auth_config=auth_config,
            )

        return await self._execute_http(
            endpoint_url, http_method, arguments,
            headers, auth_type, auth_config, timeout_sec,
        )

    async def _execute_builtin(
        self, tool, arguments: dict,
        auth_type: str = "", auth_config: dict | None = None,
    ) -> dict:
        """Execute a built-in tool locally."""
        start = time.monotonic()
        ctx = {**self.tool_context, "auth_type": auth_type, "auth_config": auth_config or {}}
        try:
            result = await tool.execute(arguments, context=ctx)
            duration_ms = int((time.monotonic() - start) * 1000)

            return {
                "result": json.dumps(result),
                "duration_ms": duration_ms,
                "error": result.get("error", "") or "",
            }
        except Exception as e:
            duration_ms = int((time.monotonic() - start) * 1000)
            logger.error("Built-in tool '%s' execution error: %s", tool.name, e)
            return {
                "result": "",
                "duration_ms": duration_ms,
                "error": str(e),
            }

    async def _execute_http(
        self,
        endpoint_url: str,
        http_method: str,
        arguments: dict,
        headers: dict | None,
        auth_type: str,
        auth_config: dict | None,
        timeout_sec: int,
    ) -> dict:
        """Execute an HTTP-based tool call."""
        start = time.monotonic()
        req_headers = dict(headers or {})
        req_headers.setdefault("Content-Type", "application/json")

        if auth_type == "bearer" and auth_config:
            token = auth_config.get("token", "")
            if token:
                req_headers["Authorization"] = f"Bearer {token}"
        elif auth_type == "api_key" and auth_config:
            key = auth_config.get("key", "")
            header_name = auth_config.get("header", "X-Api-Key")
            if key:
                req_headers[header_name] = key

        method = http_method.upper()
        timeout = aiohttp.ClientTimeout(total=timeout_sec)

        try:
            async with aiohttp.ClientSession(timeout=timeout) as session:
                kwargs: dict = {"headers": req_headers}
                if method in ("POST", "PUT", "PATCH"):
                    kwargs["json"] = arguments
                elif arguments:
                    kwargs["params"] = {k: str(v) for k, v in arguments.items()}

                async with session.request(method, endpoint_url, **kwargs) as resp:
                    duration_ms = int((time.monotonic() - start) * 1000)
                    body = await resp.text()

                    if resp.status >= 400:
                        return {
                            "result": body,
                            "duration_ms": duration_ms,
                            "error": f"HTTP {resp.status}: {body[:500]}",
                        }

                    try:
                        result = json.loads(body)
                    except json.JSONDecodeError:
                        result = body

                    return {
                        "result": json.dumps(result) if isinstance(result, (dict, list)) else str(result),
                        "duration_ms": duration_ms,
                        "error": "",
                    }

        except asyncio.TimeoutError:
            duration_ms = int((time.monotonic() - start) * 1000)
            return {"result": "", "duration_ms": duration_ms, "error": "tool call timed out"}
        except Exception as e:
            duration_ms = int((time.monotonic() - start) * 1000)
            logger.error("Tool execution error: %s", e)
            return {"result": "", "duration_ms": duration_ms, "error": str(e)}


def _parse_json(raw) -> dict:
    """Safely parse a JSON string or return the value if already a dict."""
    if isinstance(raw, dict):
        return raw
    if not raw:
        return {}
    try:
        return json.loads(raw)
    except (json.JSONDecodeError, TypeError):
        return {}


def proto_tools_to_openai_format(tools) -> list[dict]:
    """Convert proto ToolDefinition list to OpenAI tools format.

    Includes private metadata fields (_endpoint_url, _auth_config, etc.)
    used by _execute_tool() in providers/base.py for routing.
    """
    result = []
    for tool in tools:
        params = {}
        if tool.parameters_json:
            try:
                params = json.loads(tool.parameters_json)
            except json.JSONDecodeError:
                params = {"type": "object", "properties": {}}

        if "type" not in params:
            params = {"type": "object", "properties": params}

        entry = {
            "type": "function",
            "function": {
                "name": tool.name,
                "description": tool.description,
                "parameters": params,
            },
            "_endpoint_url": tool.endpoint_url,
            "_http_method": tool.http_method or "POST",
            "_headers": _parse_json(tool.headers_json),
            "_auth_type": tool.auth_type,
            "_auth_config": _parse_json(tool.auth_config_json),
            "_timeout_sec": tool.timeout_sec or 30,
            "_tool_type": tool.tool_type or "http",
            "_tool_name": tool.name,
        }
        result.append(entry)
    return result


def proto_tools_to_anthropic_format(tools) -> list[dict]:
    """Convert proto ToolDefinition list to Anthropic tools format.

    Includes private metadata fields for tool execution routing.
    """
    result = []
    for tool in tools:
        params = {}
        if tool.parameters_json:
            try:
                params = json.loads(tool.parameters_json)
            except json.JSONDecodeError:
                params = {"type": "object", "properties": {}}

        if "type" not in params:
            params = {"type": "object", "properties": params}

        entry = {
            "name": tool.name,
            "description": tool.description,
            "input_schema": params,
            "_endpoint_url": tool.endpoint_url,
            "_http_method": tool.http_method or "POST",
            "_headers": _parse_json(tool.headers_json),
            "_auth_type": tool.auth_type,
            "_auth_config": _parse_json(tool.auth_config_json),
            "_timeout_sec": tool.timeout_sec or 30,
            "_tool_type": tool.tool_type or "http",
            "_tool_name": tool.name,
        }
        result.append(entry)
    return result
