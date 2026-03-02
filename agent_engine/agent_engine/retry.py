"""Retry with exponential backoff for retryable errors."""

import asyncio
import logging
import random
from dataclasses import dataclass

from .errors import AgentEngineError, ProviderRateLimitError

logger = logging.getLogger(__name__)


@dataclass
class RetryConfig:
    max_retries: int = 3
    base_delay_sec: float = 1.0
    max_delay_sec: float = 30.0
    jitter: bool = True


async def retry_with_backoff(fn, *args, config: RetryConfig | None = None, **kwargs):
    """Call *fn* with retries on retryable AgentEngineError exceptions.

    Non-retryable errors propagate immediately.
    """
    cfg = config or RetryConfig()
    last_exc: Exception | None = None

    for attempt in range(1 + cfg.max_retries):
        try:
            return await fn(*args, **kwargs)
        except AgentEngineError as exc:
            last_exc = exc
            if not exc.retryable or attempt >= cfg.max_retries:
                raise

            # Respect provider-specified retry-after for rate limits
            if isinstance(exc, ProviderRateLimitError) and exc.retry_after_sec > 0:
                delay = min(exc.retry_after_sec, cfg.max_delay_sec)
            else:
                delay = min(cfg.base_delay_sec * (2 ** attempt), cfg.max_delay_sec)

            if cfg.jitter:
                delay *= 0.5 + random.random()  # noqa: S311

            logger.warning(
                "Retry %d/%d after %.1fs (error: %s)",
                attempt + 1,
                cfg.max_retries,
                delay,
                exc,
            )
            await asyncio.sleep(delay)

    # Should not reach here, but just in case
    if last_exc:
        raise last_exc
