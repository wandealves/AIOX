"""Tests for retry with exponential backoff."""

import asyncio
import unittest

from agent_engine.errors import (
    AgentEngineError,
    ProviderAuthError,
    ProviderRateLimitError,
    ProviderTimeoutError,
)
from agent_engine.retry import RetryConfig, retry_with_backoff


class TestRetryWithBackoff(unittest.TestCase):
    def test_success_no_retry(self):
        """Function succeeds on first call — no retries needed."""
        call_count = 0

        async def fn():
            nonlocal call_count
            call_count += 1
            return "ok"

        result = asyncio.get_event_loop().run_until_complete(
            retry_with_backoff(fn, config=RetryConfig(max_retries=3))
        )
        self.assertEqual(result, "ok")
        self.assertEqual(call_count, 1)

    def test_retry_then_success(self):
        """Fail twice with retryable error, succeed on third attempt."""
        call_count = 0

        async def fn():
            nonlocal call_count
            call_count += 1
            if call_count < 3:
                raise ProviderTimeoutError("timeout")
            return "recovered"

        cfg = RetryConfig(max_retries=3, base_delay_sec=0.01, jitter=False)
        result = asyncio.get_event_loop().run_until_complete(
            retry_with_backoff(fn, config=cfg)
        )
        self.assertEqual(result, "recovered")
        self.assertEqual(call_count, 3)

    def test_max_retries_exceeded(self):
        """All retries fail — raises the last error."""
        call_count = 0

        async def fn():
            nonlocal call_count
            call_count += 1
            raise ProviderTimeoutError("always fail")

        cfg = RetryConfig(max_retries=2, base_delay_sec=0.01, jitter=False)
        with self.assertRaises(ProviderTimeoutError):
            asyncio.get_event_loop().run_until_complete(
                retry_with_backoff(fn, config=cfg)
            )
        self.assertEqual(call_count, 3)  # 1 initial + 2 retries

    def test_non_retryable_propagates_immediately(self):
        """Non-retryable error is raised without retrying."""
        call_count = 0

        async def fn():
            nonlocal call_count
            call_count += 1
            raise ProviderAuthError("bad key")

        cfg = RetryConfig(max_retries=3, base_delay_sec=0.01)
        with self.assertRaises(ProviderAuthError):
            asyncio.get_event_loop().run_until_complete(
                retry_with_backoff(fn, config=cfg)
            )
        self.assertEqual(call_count, 1)

    def test_rate_limit_respects_retry_after(self):
        """Rate limit with retry_after_sec is respected."""
        call_count = 0

        async def fn():
            nonlocal call_count
            call_count += 1
            if call_count == 1:
                raise ProviderRateLimitError("429", retry_after_sec=0.01)
            return "ok"

        cfg = RetryConfig(max_retries=2, base_delay_sec=0.01, jitter=False)
        result = asyncio.get_event_loop().run_until_complete(
            retry_with_backoff(fn, config=cfg)
        )
        self.assertEqual(result, "ok")
        self.assertEqual(call_count, 2)

    def test_non_agent_error_propagates(self):
        """Non-AgentEngineError exceptions are not caught."""

        async def fn():
            raise ValueError("unrelated error")

        cfg = RetryConfig(max_retries=3, base_delay_sec=0.01)
        with self.assertRaises(ValueError):
            asyncio.get_event_loop().run_until_complete(
                retry_with_backoff(fn, config=cfg)
            )


if __name__ == "__main__":
    unittest.main()
