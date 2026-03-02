"""Execution budget tracking — limits steps, tokens, time, and cost per task."""

import time
from dataclasses import dataclass, field

from .cost import estimate_cost
from .errors import BudgetExceededError


@dataclass
class ExecutionBudget:
    max_steps: int = 15
    max_tokens: int = 50_000
    max_time_sec: float = 90.0
    max_cost_usd: float = 1.0


class BudgetTracker:
    """Tracks resource usage during a single task execution and raises
    BudgetExceededError when any limit is reached."""

    def __init__(self, budget: ExecutionBudget):
        self.budget = budget
        self._steps: int = 0
        self._tokens: int = 0
        self._cost: float = 0.0
        self._start: float = field(default_factory=time.monotonic)  # type: ignore[assignment]
        self._start = time.monotonic()

    # -- accessors ----------------------------------------------------------

    @property
    def steps(self) -> int:
        return self._steps

    @property
    def tokens(self) -> int:
        return self._tokens

    @property
    def cost(self) -> float:
        return self._cost

    @property
    def elapsed_sec(self) -> float:
        return time.monotonic() - self._start

    # -- mutation -----------------------------------------------------------

    def record_step(self) -> None:
        self._steps += 1

    def record_tokens(self, tokens: int, model: str = "") -> None:
        self._tokens += tokens
        if model:
            self._cost += estimate_cost(model, tokens)

    # -- checks -------------------------------------------------------------

    def check_steps(self) -> None:
        if self._steps >= self.budget.max_steps:
            raise BudgetExceededError(
                f"Step limit reached ({self._steps}/{self.budget.max_steps})",
                budget_type="steps",
                limit=self.budget.max_steps,
                current=self._steps,
            )

    def check_tokens(self) -> None:
        if self._tokens >= self.budget.max_tokens:
            raise BudgetExceededError(
                f"Token limit reached ({self._tokens}/{self.budget.max_tokens})",
                budget_type="tokens",
                limit=self.budget.max_tokens,
                current=self._tokens,
            )

    def check_time(self) -> None:
        elapsed = self.elapsed_sec
        if elapsed >= self.budget.max_time_sec:
            raise BudgetExceededError(
                f"Time limit reached ({elapsed:.1f}s/{self.budget.max_time_sec}s)",
                budget_type="time",
                limit=self.budget.max_time_sec,
                current=elapsed,
            )

    def check_cost(self) -> None:
        if self._cost >= self.budget.max_cost_usd:
            raise BudgetExceededError(
                f"Cost limit reached (${self._cost:.4f}/${self.budget.max_cost_usd})",
                budget_type="cost",
                limit=self.budget.max_cost_usd,
                current=self._cost,
            )

    def check_all(self, new_tokens: int = 0, model: str = "") -> None:
        """Record new tokens and check every limit at once."""
        if new_tokens:
            self.record_tokens(new_tokens, model)
        self.check_steps()
        self.check_tokens()
        self.check_time()
        self.check_cost()
