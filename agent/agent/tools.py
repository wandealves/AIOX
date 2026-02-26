"""Tool executor for agent function calling."""
import json
import logging
import time

import aiohttp

logger = logging.getLogger(__name__)


class ToolExecutor:
    """Executes HTTP-based tool calls for LLM function calling."""

    async def execute(
        self,
        endpoint_url: str,
        http_method: str,
        arguments: dict,
        headers: dict | None = None,
        auth_type: str = "",
        auth_config: dict | None = None,
        timeout_sec: int = 30,
    ) -> dict:
        """Execute an HTTP tool call and return the result.

        Returns dict with keys: result, duration_ms, error
        """
        start = time.monotonic()
        req_headers = dict(headers or {})
        req_headers.setdefault("Content-Type", "application/json")

        # Apply auth
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
                kwargs = {"headers": req_headers}
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

                    # Try to parse as JSON
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


# Need asyncio for TimeoutError
import asyncio


def proto_tools_to_openai_format(tools):
    """Convert proto ToolDefinition list to OpenAI tools format."""
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

        result.append({
            "type": "function",
            "function": {
                "name": tool.name,
                "description": tool.description,
                "parameters": params,
            },
        })
    return result


def proto_tools_to_anthropic_format(tools):
    """Convert proto ToolDefinition list to Anthropic tools format."""
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

        result.append({
            "name": tool.name,
            "description": tool.description,
            "input_schema": params,
        })
    return result
