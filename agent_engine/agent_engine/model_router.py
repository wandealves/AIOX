"""Model router — selects the appropriate model tier based on task complexity."""

import re

# Model tiers per provider: small (fast/cheap), medium (balanced), large (powerful)
MODEL_TIERS: dict[str, dict[str, str]] = {
    "openai": {
        "small": "gpt-4o-mini",
        "medium": "gpt-4o-mini",
        "large": "gpt-4o",
    },
    "anthropic": {
        "small": "claude-haiku-4-20250514",
        "medium": "claude-sonnet-4-20250514",
        "large": "claude-opus-4-20250514",
    },
    "gemini": {
        "small": "gemini-2.0-flash",
        "medium": "gemini-2.0-flash",
        "large": "gemini-1.5-pro",
    },
    "deepseek": {
        "small": "deepseek-chat",
        "medium": "deepseek-chat",
        "large": "deepseek-reasoner",
    },
    "ollama": {
        "small": "llama3.2",
        "medium": "llama3.2",
        "large": "llama3.2",
    },
}

_COMPLEXITY_KEYWORDS = re.compile(
    r"(analyz|compar|evaluat|summar|explain\s+in\s+detail|write\s+a\s+(long|detailed)"
    r"|research|investigate|create\s+a\s+plan|step.by.step|comprehensive)",
    re.IGNORECASE,
)


class ComplexityEstimator:
    """Heuristic complexity estimation based on message characteristics."""

    @staticmethod
    def estimate(
        user_message: str,
        has_tools: bool = False,
        num_tools: int = 0,
        has_memory: bool = False,
    ) -> str:
        """Return 'small', 'medium', or 'large'."""
        score = 0

        # Message length
        msg_len = len(user_message)
        if msg_len > 2000:
            score += 3
        elif msg_len > 500:
            score += 1

        # Tool usage
        if has_tools:
            score += 1
            if num_tools > 3:
                score += 2

        # Memory context
        if has_memory:
            score += 1

        # Complexity keywords
        if _COMPLEXITY_KEYWORDS.search(user_message):
            score += 2

        if score >= 5:
            return "large"
        if score >= 2:
            return "medium"
        return "small"


def route_model(
    provider: str,
    user_message: str,
    has_tools: bool = False,
    num_tools: int = 0,
    has_memory: bool = False,
    explicit_model: str = "",
) -> str:
    """Select the best model for the given context.

    If *explicit_model* is set, it takes precedence.
    """
    if explicit_model:
        return explicit_model

    tier = ComplexityEstimator.estimate(
        user_message, has_tools, num_tools, has_memory,
    )

    provider_tiers = MODEL_TIERS.get(provider, MODEL_TIERS.get("openai", {}))
    return provider_tiers.get(tier, provider_tiers.get("medium", ""))
