"""Tests for UTCP tool type handling in executor."""

import asyncio
import logging
import unittest
from unittest.mock import AsyncMock, patch

from agent_engine.tools.executor import ToolExecutor


class TestUTCPExecutor(unittest.TestCase):
    def test_utcp_tool_logging(self):
        """UTCP tool execution should log the tool type."""
        executor = ToolExecutor()

        with self.assertLogs("agent_engine.tools.executor", level="INFO") as cm:
            asyncio.get_event_loop().run_until_complete(
                executor.execute(
                    endpoint_url="http://localhost:1/nonexistent",
                    http_method="GET",
                    arguments={},
                    tool_type="utcp",
                    tool_name="test.tool",
                )
            )

        utcp_logs = [msg for msg in cm.output if "UTCP tool" in msg]
        self.assertEqual(len(utcp_logs), 1)
        self.assertIn("test.tool", utcp_logs[0])

    def test_non_utcp_tool_no_utcp_log(self):
        """Regular HTTP tools should not emit UTCP-specific log."""
        executor = ToolExecutor()

        # Patch logger to capture calls
        with patch.object(logging.getLogger("agent_engine.tools.executor"), "info") as mock_info:
            asyncio.get_event_loop().run_until_complete(
                executor.execute(
                    endpoint_url="http://localhost:1/nonexistent",
                    http_method="GET",
                    arguments={},
                    tool_type="http",
                    tool_name="regular.tool",
                )
            )

        utcp_calls = [c for c in mock_info.call_args_list if "UTCP tool" in str(c)]
        self.assertEqual(len(utcp_calls), 0)

    def test_utcp_tool_routes_to_http(self):
        """UTCP tools should fall through to _execute_http since they're not builtins."""
        executor = ToolExecutor()
        executor._execute_http = AsyncMock(return_value={
            "result": '{"ok": true}',
            "duration_ms": 10,
            "error": "",
        })

        result = asyncio.get_event_loop().run_until_complete(
            executor.execute(
                endpoint_url="http://example.com/api",
                http_method="POST",
                arguments={"q": "test"},
                tool_type="utcp",
                tool_name="weather.search",
            )
        )

        executor._execute_http.assert_called_once()
        self.assertEqual(result["error"], "")
        self.assertIn("ok", result["result"])


    def test_session_token_capture_from_login(self):
        """Token from a login response should be captured and reused."""
        executor = ToolExecutor()

        # Simulate login returning a token
        executor._execute_http = AsyncMock(return_value={
            "result": '{"token": "jwt-abc-123"}',
            "duration_ms": 50,
            "error": "",
        })

        asyncio.get_event_loop().run_until_complete(
            executor.execute(
                endpoint_url="http://api.com/auth/login",
                http_method="POST",
                arguments={"username": "u", "password": "p"},
                tool_type="utcp",
                tool_name="MyApi__login",
            )
        )

        # Token should be captured under the "MyApi" namespace
        self.assertEqual(executor._session_tokens.get("MyApi"), "jwt-abc-123")

    def test_session_token_injected_into_next_call(self):
        """Captured token should be injected into subsequent UTCP calls."""
        executor = ToolExecutor()
        executor._session_tokens["MyApi"] = "jwt-abc-123"

        executor._execute_http = AsyncMock(return_value={
            "result": '[]',
            "duration_ms": 10,
            "error": "",
        })

        asyncio.get_event_loop().run_until_complete(
            executor.execute(
                endpoint_url="http://api.com/products",
                http_method="GET",
                arguments={},
                auth_type="bearer",
                auth_config={"token": ""},  # empty static token
                tool_type="utcp",
                tool_name="MyApi__list_products",
            )
        )

        # Verify the injected auth_config has the session token
        call_kwargs = executor._execute_http.call_args
        self.assertEqual(call_kwargs.kwargs.get("auth_config") or call_kwargs[0][5], {"token": "jwt-abc-123"})

    def test_static_token_not_overridden(self):
        """If a static token is provided, session token should NOT override it."""
        executor = ToolExecutor()
        executor._session_tokens["MyApi"] = "session-token"

        executor._execute_http = AsyncMock(return_value={
            "result": '[]',
            "duration_ms": 10,
            "error": "",
        })

        asyncio.get_event_loop().run_until_complete(
            executor.execute(
                endpoint_url="http://api.com/products",
                http_method="GET",
                arguments={},
                auth_type="bearer",
                auth_config={"token": "static-token"},
                tool_type="utcp",
                tool_name="MyApi__list_products",
            )
        )

        call_kwargs = executor._execute_http.call_args
        auth = call_kwargs.kwargs.get("auth_config") or call_kwargs[0][5]
        self.assertEqual(auth["token"], "static-token")


if __name__ == "__main__":
    unittest.main()
