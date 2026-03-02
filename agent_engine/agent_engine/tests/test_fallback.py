"""Tests for fallback chain."""

import asyncio
import unittest
from unittest.mock import AsyncMock, MagicMock

from agent_engine.errors import (
    AgentEngineError,
    ProviderAuthError,
    ProviderTimeoutError,
)
from agent_engine.fallback import FallbackChain
from agent_engine.providers.base import LLMResponse


def _ok_response(text="ok"):
    return LLMResponse(text=text, tokens_used=10, model_used="test", duration_ms=100)


class TestFallbackChain(unittest.TestCase):
    def test_primary_succeeds(self):
        p1 = MagicMock()
        p1.generate = AsyncMock(return_value=_ok_response("from p1"))

        chain = FallbackChain([("p1", p1)])
        result = asyncio.get_event_loop().run_until_complete(
            chain.generate(
                system_prompt="sys",
                user_message="hi",
                model="",
                temperature=0.7,
                max_tokens=100,
            )
        )
        self.assertEqual(result.text, "from p1")

    def test_fallback_after_retryable_error(self):
        p1 = MagicMock()
        p1.generate = AsyncMock(side_effect=ProviderTimeoutError("timeout"))
        p2 = MagicMock()
        p2.generate = AsyncMock(return_value=_ok_response("from p2"))

        chain = FallbackChain([("p1", p1), ("p2", p2)])
        result = asyncio.get_event_loop().run_until_complete(
            chain.generate(
                system_prompt="sys",
                user_message="hi",
                model="",
                temperature=0.7,
                max_tokens=100,
            )
        )
        self.assertEqual(result.text, "from p2")

    def test_non_retryable_propagates(self):
        p1 = MagicMock()
        p1.generate = AsyncMock(side_effect=ProviderAuthError("bad key"))
        p2 = MagicMock()
        p2.generate = AsyncMock(return_value=_ok_response())

        chain = FallbackChain([("p1", p1), ("p2", p2)])
        with self.assertRaises(ProviderAuthError):
            asyncio.get_event_loop().run_until_complete(
                chain.generate(
                    system_prompt="sys",
                    user_message="hi",
                    model="",
                    temperature=0.7,
                    max_tokens=100,
                )
            )
        p2.generate.assert_not_called()

    def test_all_fail(self):
        p1 = MagicMock()
        p1.generate = AsyncMock(side_effect=ProviderTimeoutError("t1"))
        p2 = MagicMock()
        p2.generate = AsyncMock(side_effect=ProviderTimeoutError("t2"))

        chain = FallbackChain([("p1", p1), ("p2", p2)])
        with self.assertRaises(ProviderTimeoutError):
            asyncio.get_event_loop().run_until_complete(
                chain.generate(
                    system_prompt="sys",
                    user_message="hi",
                    model="",
                    temperature=0.7,
                    max_tokens=100,
                )
            )


if __name__ == "__main__":
    unittest.main()
