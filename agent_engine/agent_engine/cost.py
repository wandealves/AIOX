"""Per-model cost estimation for budget tracking."""

# Cost per 1K tokens (blended input+output average).
# Values are approximate and should be updated as providers change pricing.
MODEL_COSTS: dict[str, float] = {
    # OpenAI
    "gpt-4o": 0.005,
    "gpt-4o-mini": 0.00015,
    "gpt-4-turbo": 0.01,
    "gpt-4": 0.03,
    "gpt-3.5-turbo": 0.0005,
    # Anthropic
    "claude-opus-4-20250514": 0.015,
    "claude-sonnet-4-20250514": 0.003,
    "claude-haiku-4-20250514": 0.0008,
    # Google
    "gemini-2.0-flash": 0.0001,
    "gemini-1.5-pro": 0.00125,
    "gemini-1.5-flash": 0.0001,
    # DeepSeek
    "deepseek-chat": 0.00014,
    "deepseek-reasoner": 0.00055,
}

# Fallback cost when model is unknown
_DEFAULT_COST_PER_1K = 0.001


def estimate_cost(model: str, tokens: int) -> float:
    """Estimate USD cost for *tokens* on *model*."""
    cost_per_1k = MODEL_COSTS.get(model, _DEFAULT_COST_PER_1K)
    return (tokens / 1000.0) * cost_per_1k
