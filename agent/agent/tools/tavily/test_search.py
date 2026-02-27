"""Tests for the Tavily Search tool."""

import asyncio
import json
import os
import unittest
from unittest.mock import AsyncMock, MagicMock, patch

from .search import TAVILY_SEARCH_PARAMETERS, TavilySearchTool


class TestTavilySearchTool(unittest.TestCase):
    def setUp(self):
        self.tool = TavilySearchTool()

    def _run(self, coro):
        return asyncio.get_event_loop().run_until_complete(coro)

    def test_name_and_description(self):
        self.assertEqual(self.tool.name, "tavily_search")
        self.assertIn("search", self.tool.description.lower())
        self.assertTrue(len(self.tool.description) > 10)

    def test_missing_api_key(self):
        with patch.dict(os.environ, {}, clear=True):
            os.environ.pop("TAVILY_API_KEY", None)
            result = self._run(self.tool.execute({"query": "test"}))

        self.assertEqual(result["status"], "error")
        self.assertIn("not configured", result["error"])
        self.assertIsNone(result["data"])

    def test_missing_query(self):
        with patch.dict(os.environ, {"TAVILY_API_KEY": "tvly-test-key"}):
            result = self._run(self.tool.execute({}))

        self.assertEqual(result["status"], "error")
        self.assertIn("query", result["error"])

    def test_empty_query(self):
        with patch.dict(os.environ, {"TAVILY_API_KEY": "tvly-test-key"}):
            result = self._run(self.tool.execute({"query": "   "}))

        self.assertEqual(result["status"], "error")
        self.assertIn("query", result["error"])

    @patch("agent.tools.tavily.search.aiohttp.ClientSession")
    def test_successful_search(self, mock_session_cls):
        api_response = {
            "query": "python asyncio",
            "answer": "Python asyncio is a library for async programming.",
            "results": [
                {
                    "title": "asyncio docs",
                    "url": "https://docs.python.org/3/library/asyncio.html",
                    "content": "asyncio is a library to write concurrent code.",
                    "score": 0.95,
                },
                {
                    "title": "Real Python",
                    "url": "https://realpython.com/async-io-python/",
                    "content": "Async IO in Python: A Complete Walkthrough.",
                    "score": 0.88,
                },
            ],
        }

        mock_resp = AsyncMock()
        mock_resp.status = 200
        mock_resp.text = AsyncMock(return_value=json.dumps(api_response))

        mock_ctx = AsyncMock()
        mock_ctx.__aenter__ = AsyncMock(return_value=mock_resp)
        mock_ctx.__aexit__ = AsyncMock(return_value=False)

        mock_session = AsyncMock()
        mock_session.post = MagicMock(return_value=mock_ctx)
        mock_session.__aenter__ = AsyncMock(return_value=mock_session)
        mock_session.__aexit__ = AsyncMock(return_value=False)

        mock_session_cls.return_value = mock_session

        with patch.dict(os.environ, {"TAVILY_API_KEY": "tvly-test-key"}):
            result = self._run(
                self.tool.execute({
                    "query": "python asyncio",
                    "max_results": 2,
                    "include_answer": True,
                })
            )

        self.assertEqual(result["status"], "success")
        self.assertIsNone(result["error"])
        self.assertEqual(result["data"]["query"], "python asyncio")
        self.assertEqual(len(result["data"]["results"]), 2)
        self.assertEqual(result["data"]["results"][0]["title"], "asyncio docs")
        self.assertIn("asyncio", result["data"]["answer"])

    @patch("agent.tools.tavily.search.aiohttp.ClientSession")
    def test_api_error_response(self, mock_session_cls):
        mock_resp = AsyncMock()
        mock_resp.status = 401
        mock_resp.text = AsyncMock(return_value='{"error": "Invalid API key"}')

        mock_ctx = AsyncMock()
        mock_ctx.__aenter__ = AsyncMock(return_value=mock_resp)
        mock_ctx.__aexit__ = AsyncMock(return_value=False)

        mock_session = AsyncMock()
        mock_session.post = MagicMock(return_value=mock_ctx)
        mock_session.__aenter__ = AsyncMock(return_value=mock_session)
        mock_session.__aexit__ = AsyncMock(return_value=False)

        mock_session_cls.return_value = mock_session

        with patch.dict(os.environ, {"TAVILY_API_KEY": "tvly-bad-key"}):
            result = self._run(self.tool.execute({"query": "test"}))

        self.assertEqual(result["status"], "error")
        self.assertIn("401", result["error"])

    @patch("agent.tools.tavily.search.aiohttp.ClientSession")
    def test_server_error_response(self, mock_session_cls):
        mock_resp = AsyncMock()
        mock_resp.status = 500
        mock_resp.text = AsyncMock(return_value="Internal Server Error")

        mock_ctx = AsyncMock()
        mock_ctx.__aenter__ = AsyncMock(return_value=mock_resp)
        mock_ctx.__aexit__ = AsyncMock(return_value=False)

        mock_session = AsyncMock()
        mock_session.post = MagicMock(return_value=mock_ctx)
        mock_session.__aenter__ = AsyncMock(return_value=mock_session)
        mock_session.__aexit__ = AsyncMock(return_value=False)

        mock_session_cls.return_value = mock_session

        with patch.dict(os.environ, {"TAVILY_API_KEY": "tvly-test-key"}):
            result = self._run(self.tool.execute({"query": "test"}))

        self.assertEqual(result["status"], "error")
        self.assertIn("500", result["error"])

    def test_parameters_schema(self):
        self.assertEqual(TAVILY_SEARCH_PARAMETERS["type"], "object")
        self.assertIn("query", TAVILY_SEARCH_PARAMETERS["properties"])
        self.assertIn("search_depth", TAVILY_SEARCH_PARAMETERS["properties"])
        self.assertIn("max_results", TAVILY_SEARCH_PARAMETERS["properties"])
        self.assertIn("include_answer", TAVILY_SEARCH_PARAMETERS["properties"])
        self.assertEqual(TAVILY_SEARCH_PARAMETERS["required"], ["query"])
        self.assertEqual(
            TAVILY_SEARCH_PARAMETERS["properties"]["query"]["type"], "string"
        )

    def test_registered_in_builtins(self):
        # Importing the tools package triggers registration
        from ..registry import get_tool as get_builtin_tool, is_builtin as is_builtin_tool

        self.assertTrue(is_builtin_tool("tavily_search"))
        tool = get_builtin_tool("tavily_search")
        self.assertIsNotNone(tool)
        self.assertIsInstance(tool, TavilySearchTool)

    @patch("agent.tools.tavily.search.aiohttp.ClientSession")
    def test_api_key_from_arguments(self, mock_session_cls):
        """API key should be resolved from arguments (parameter default)."""
        api_response = {"results": [], "answer": ""}
        mock_resp = AsyncMock()
        mock_resp.status = 200
        mock_resp.text = AsyncMock(return_value=json.dumps(api_response))

        mock_ctx = AsyncMock()
        mock_ctx.__aenter__ = AsyncMock(return_value=mock_resp)
        mock_ctx.__aexit__ = AsyncMock(return_value=False)

        mock_session = AsyncMock()
        mock_session.post = MagicMock(return_value=mock_ctx)
        mock_session.__aenter__ = AsyncMock(return_value=mock_session)
        mock_session.__aexit__ = AsyncMock(return_value=False)

        mock_session_cls.return_value = mock_session

        with patch.dict(os.environ, {}, clear=True):
            os.environ.pop("TAVILY_API_KEY", None)
            result = self._run(
                self.tool.execute({"query": "test", "api_key": "tvly-from-args"})
            )

        self.assertEqual(result["status"], "success")
        call_args = mock_session.post.call_args
        payload = call_args[1]["json"]
        self.assertEqual(payload["api_key"], "tvly-from-args")

    @patch("agent.tools.tavily.search.aiohttp.ClientSession")
    def test_api_key_from_auth_config(self, mock_session_cls):
        """API key should be resolved from context auth_config instead of env var."""
        api_response = {"results": [], "answer": ""}
        mock_resp = AsyncMock()
        mock_resp.status = 200
        mock_resp.text = AsyncMock(return_value=json.dumps(api_response))

        mock_ctx = AsyncMock()
        mock_ctx.__aenter__ = AsyncMock(return_value=mock_resp)
        mock_ctx.__aexit__ = AsyncMock(return_value=False)

        mock_session = AsyncMock()
        mock_session.post = MagicMock(return_value=mock_ctx)
        mock_session.__aenter__ = AsyncMock(return_value=mock_session)
        mock_session.__aexit__ = AsyncMock(return_value=False)

        mock_session_cls.return_value = mock_session

        context = {"auth_config": {"key": "tvly-from-config", "header": "X-Api-Key"}}

        with patch.dict(os.environ, {}, clear=True):
            os.environ.pop("TAVILY_API_KEY", None)
            result = self._run(self.tool.execute({"query": "test"}, context=context))

        self.assertEqual(result["status"], "success")
        # Verify the payload used the key from auth_config
        call_args = mock_session.post.call_args
        payload = call_args[1]["json"]
        self.assertEqual(payload["api_key"], "tvly-from-config")

    @patch("agent.tools.tavily.search.aiohttp.ClientSession")
    def test_api_key_from_bearer_token(self, mock_session_cls):
        """API key should be resolved from context auth_config bearer token."""
        api_response = {"results": [], "answer": ""}
        mock_resp = AsyncMock()
        mock_resp.status = 200
        mock_resp.text = AsyncMock(return_value=json.dumps(api_response))

        mock_ctx = AsyncMock()
        mock_ctx.__aenter__ = AsyncMock(return_value=mock_resp)
        mock_ctx.__aexit__ = AsyncMock(return_value=False)

        mock_session = AsyncMock()
        mock_session.post = MagicMock(return_value=mock_ctx)
        mock_session.__aenter__ = AsyncMock(return_value=mock_session)
        mock_session.__aexit__ = AsyncMock(return_value=False)

        mock_session_cls.return_value = mock_session

        context = {"auth_config": {"token": "tvly-bearer-key"}}

        with patch.dict(os.environ, {}, clear=True):
            os.environ.pop("TAVILY_API_KEY", None)
            result = self._run(self.tool.execute({"query": "test"}, context=context))

        self.assertEqual(result["status"], "success")
        call_args = mock_session.post.call_args
        payload = call_args[1]["json"]
        self.assertEqual(payload["api_key"], "tvly-bearer-key")

    def test_invalid_search_depth_defaults_to_basic(self):
        """Invalid search_depth should fall back to 'basic'."""
        with patch.dict(os.environ, {"TAVILY_API_KEY": "tvly-test-key"}):
            with patch("agent.tools.tavily.search.aiohttp.ClientSession") as mock_cls:
                api_response = {"results": [], "answer": ""}
                mock_resp = AsyncMock()
                mock_resp.status = 200
                mock_resp.text = AsyncMock(return_value=json.dumps(api_response))

                mock_ctx = AsyncMock()
                mock_ctx.__aenter__ = AsyncMock(return_value=mock_resp)
                mock_ctx.__aexit__ = AsyncMock(return_value=False)

                mock_session = AsyncMock()
                mock_session.post = MagicMock(return_value=mock_ctx)
                mock_session.__aenter__ = AsyncMock(return_value=mock_session)
                mock_session.__aexit__ = AsyncMock(return_value=False)

                mock_cls.return_value = mock_session

                result = self._run(
                    self.tool.execute({
                        "query": "test",
                        "search_depth": "invalid",
                    })
                )

                self.assertEqual(result["status"], "success")
                # Verify the payload used 'basic'
                call_args = mock_session.post.call_args
                payload = call_args[1]["json"]
                self.assertEqual(payload["search_depth"], "basic")


if __name__ == "__main__":
    unittest.main()
