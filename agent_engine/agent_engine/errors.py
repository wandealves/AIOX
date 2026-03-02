"""Structured error hierarchy for the agent engine.

All errors inherit from AgentEngineError and carry a `retryable` flag
so the retry/fallback layers can decide whether to re-attempt the call.
"""


class AgentEngineError(Exception):
    """Base exception for all agent engine errors."""

    def __init__(self, message: str, *, retryable: bool = False):
        super().__init__(message)
        self.retryable = retryable


# ---------------------------------------------------------------------------
# Provider errors
# ---------------------------------------------------------------------------

class ProviderError(AgentEngineError):
    """An LLM provider returned an error."""


class ProviderRateLimitError(ProviderError):
    """429 — rate limit exceeded."""

    def __init__(self, message: str, *, retry_after_sec: float = 0.0):
        super().__init__(message, retryable=True)
        self.retry_after_sec = retry_after_sec


class ProviderAuthError(ProviderError):
    """401/403 — authentication or authorization failure."""

    def __init__(self, message: str):
        super().__init__(message, retryable=False)


class ProviderOverloadedError(ProviderError):
    """503/529 — provider is temporarily overloaded."""

    def __init__(self, message: str):
        super().__init__(message, retryable=True)


class ProviderTimeoutError(ProviderError):
    """Request to the provider timed out."""

    def __init__(self, message: str):
        super().__init__(message, retryable=True)


class ProviderContentFilterError(ProviderError):
    """Content was blocked by the provider's safety filters."""

    def __init__(self, message: str):
        super().__init__(message, retryable=False)


# ---------------------------------------------------------------------------
# Budget errors
# ---------------------------------------------------------------------------

class BudgetExceededError(AgentEngineError):
    """An execution budget limit was reached."""

    def __init__(
        self,
        message: str,
        *,
        budget_type: str = "",
        limit: float = 0.0,
        current: float = 0.0,
    ):
        super().__init__(message, retryable=False)
        self.budget_type = budget_type
        self.limit = limit
        self.current = current


# ---------------------------------------------------------------------------
# Tool errors
# ---------------------------------------------------------------------------

class ToolExecutionError(AgentEngineError):
    """A tool call failed during execution."""


class ToolTimeoutError(ToolExecutionError):
    """A tool call timed out."""

    def __init__(self, message: str):
        super().__init__(message, retryable=True)
