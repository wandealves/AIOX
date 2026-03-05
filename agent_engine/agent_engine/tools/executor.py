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

    For UTCP tools with dynamic auth (e.g. login returns a JWT),
    captured tokens are automatically injected into subsequent
    UTCP tool calls within the same session.
    """

    def __init__(self, tool_context: dict | None = None, policy=None):
        self.tool_context = tool_context or {}
        self.policy = policy  # Optional Policy instance for domain filtering
        self._session_tokens: dict[str, str] = {}  # namespace -> bearer token

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
                builtin,
                arguments,
                auth_type=auth_type,
                auth_config=auth_config,
            )

        if tool_type == "utcp":
            logger.info(
                "Executing UTCP tool: %s via %s %s",
                tool_name, http_method, endpoint_url,
            )
            # Inject session token if available and no static token was provided
            namespace = tool_name.split("__")[0] if "__" in tool_name else ""
            if namespace and auth_type == "bearer":
                cfg = dict(auth_config or {})
                if not cfg.get("token") and namespace in self._session_tokens:
                    cfg["token"] = self._session_tokens[namespace]
                    auth_config = cfg

        result = await self._execute_http(
            endpoint_url,
            http_method,
            arguments,
            headers,
            auth_type,
            auth_config,
            timeout_sec,
        )

        # Capture bearer tokens from UTCP tool responses (e.g. login)
        if tool_type == "utcp" and not result.get("error"):
            self._capture_session_token(tool_name, result.get("result", ""))

        return result

    def _capture_session_token(self, tool_name: str, raw_result: str) -> None:
        """Capture a bearer token from a UTCP tool response.

        When a tool (e.g. login) returns a JSON object with a "token" field,
        store it so subsequent UTCP calls in the same namespace can use it.
        """
        if not raw_result:
            return
        try:
            data = json.loads(raw_result)
        except (json.JSONDecodeError, TypeError):
            return
        if not isinstance(data, dict):
            return
        token = data.get("token") or data.get("access_token") or ""
        if token and isinstance(token, str):
            namespace = tool_name.split("__")[0] if "__" in tool_name else ""
            if namespace:
                self._session_tokens[namespace] = token
                logger.info(
                    "Captured session token from %s for namespace %s",
                    tool_name, namespace,
                )

    async def _execute_builtin(
        self,
        tool,
        arguments: dict,
        auth_type: str = "",
        auth_config: dict | None = None,
    ) -> dict:
        """Execute a built-in tool locally."""
        start = time.monotonic()
        ctx = {
            **self.tool_context,
            "auth_type": auth_type,
            "auth_config": auth_config or {},
        }
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
        # Policy: check URL domain
        if self.policy and endpoint_url:
            if not self.policy.check_tool_url(endpoint_url):
                return {
                    "result": "",
                    "duration_ms": 0,
                    "error": f"Domain not allowed by policy: {endpoint_url}",
                }
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

                    raw_result = (
                        json.dumps(result)
                        if isinstance(result, (dict, list))
                        else str(result)
                    )
                    try:
                        from ..sanitize import sanitize_tool_result

                        raw_result = sanitize_tool_result(raw_result)
                    except ImportError:
                        pass
                    return {
                        "result": raw_result,
                        "duration_ms": duration_ms,
                        "error": "",
                    }

        except asyncio.TimeoutError:
            duration_ms = int((time.monotonic() - start) * 1000)
            return {
                "result": "",
                "duration_ms": duration_ms,
                "error": "tool call timed out",
            }
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
