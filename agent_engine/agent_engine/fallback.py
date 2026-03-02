"""Fallback chain — tries multiple providers in order on retryable failures."""

import logging
from typing import TYPE_CHECKING

from .errors import AgentEngineError

if TYPE_CHECKING:
    from .providers.base import LLMProvider, LLMResponse

logger = logging.getLogger(__name__)


class FallbackChain:
    """Tries each provider in order. Non-retryable errors propagate immediately."""

    def __init__(self, providers: list[tuple[str, "LLMProvider"]]):
        self._providers = providers

    async def generate(self, **kwargs) -> "LLMResponse":
        last_exc: Exception | None = None

        for name, provider in self._providers:
            try:
                logger.info("Fallback: trying provider %s", name)
                return await provider.generate(**kwargs)
            except AgentEngineError as e:
                if not e.retryable:
                    logger.warning("Fallback: %s non-retryable error, propagating: %s", name, e)
                    raise
                logger.warning("Fallback: %s failed (retryable): %s", name, e)
                last_exc = e
            except Exception as e:
                logger.warning("Fallback: %s unexpected error: %s", name, e)
                last_exc = e

        # All providers failed
        if last_exc:
            raise last_exc
        raise AgentEngineError("No providers available in fallback chain")
