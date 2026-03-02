"""Tests for the error hierarchy."""

import unittest

from agent_engine.errors import (
    AgentEngineError,
    BudgetExceededError,
    ProviderAuthError,
    ProviderContentFilterError,
    ProviderError,
    ProviderOverloadedError,
    ProviderRateLimitError,
    ProviderTimeoutError,
    ToolExecutionError,
    ToolTimeoutError,
)


class TestErrorHierarchy(unittest.TestCase):
    def test_base_error_default_not_retryable(self):
        err = AgentEngineError("boom")
        self.assertFalse(err.retryable)
        self.assertEqual(str(err), "boom")

    def test_base_error_retryable(self):
        err = AgentEngineError("retry me", retryable=True)
        self.assertTrue(err.retryable)

    def test_provider_error_is_agent_engine_error(self):
        err = ProviderError("provider issue")
        self.assertIsInstance(err, AgentEngineError)

    def test_rate_limit_retryable(self):
        err = ProviderRateLimitError("429", retry_after_sec=5.0)
        self.assertTrue(err.retryable)
        self.assertEqual(err.retry_after_sec, 5.0)

    def test_auth_error_not_retryable(self):
        err = ProviderAuthError("401 bad key")
        self.assertFalse(err.retryable)
        self.assertIsInstance(err, ProviderError)

    def test_overloaded_retryable(self):
        err = ProviderOverloadedError("503")
        self.assertTrue(err.retryable)

    def test_timeout_retryable(self):
        err = ProviderTimeoutError("timed out")
        self.assertTrue(err.retryable)

    def test_content_filter_not_retryable(self):
        err = ProviderContentFilterError("blocked")
        self.assertFalse(err.retryable)

    def test_budget_exceeded_fields(self):
        err = BudgetExceededError(
            "too many tokens",
            budget_type="tokens",
            limit=50000,
            current=50001,
        )
        self.assertFalse(err.retryable)
        self.assertEqual(err.budget_type, "tokens")
        self.assertEqual(err.limit, 50000)
        self.assertEqual(err.current, 50001)

    def test_tool_execution_error(self):
        err = ToolExecutionError("failed")
        self.assertIsInstance(err, AgentEngineError)

    def test_tool_timeout_retryable(self):
        err = ToolTimeoutError("tool timed out")
        self.assertTrue(err.retryable)
        self.assertIsInstance(err, ToolExecutionError)

    def test_inheritance_chain(self):
        err = ProviderRateLimitError("429")
        self.assertIsInstance(err, ProviderError)
        self.assertIsInstance(err, AgentEngineError)
        self.assertIsInstance(err, Exception)


if __name__ == "__main__":
    unittest.main()
