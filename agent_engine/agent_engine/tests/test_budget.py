"""Tests for budget tracking and cost estimation."""

import time
import unittest

from agent_engine.budget import BudgetTracker, ExecutionBudget
from agent_engine.cost import MODEL_COSTS, estimate_cost
from agent_engine.errors import BudgetExceededError


class TestEstimateCost(unittest.TestCase):
    def test_known_model(self):
        cost = estimate_cost("gpt-4o-mini", 1000)
        self.assertAlmostEqual(cost, MODEL_COSTS["gpt-4o-mini"])

    def test_unknown_model_uses_default(self):
        cost = estimate_cost("unknown-model-xyz", 1000)
        self.assertAlmostEqual(cost, 0.001)

    def test_zero_tokens(self):
        self.assertEqual(estimate_cost("gpt-4o", 0), 0.0)


class TestBudgetTracker(unittest.TestCase):
    def test_within_budget(self):
        budget = ExecutionBudget(
            max_steps=10, max_tokens=10000, max_time_sec=60, max_cost_usd=1.0
        )
        tracker = BudgetTracker(budget)
        tracker.record_step()
        tracker.record_tokens(100, "gpt-4o-mini")
        # Should not raise
        tracker.check_all()

    def test_step_limit(self):
        budget = ExecutionBudget(max_steps=2)
        tracker = BudgetTracker(budget)
        tracker.record_step()
        tracker.record_step()
        with self.assertRaises(BudgetExceededError) as ctx:
            tracker.check_steps()
        self.assertEqual(ctx.exception.budget_type, "steps")

    def test_token_limit(self):
        budget = ExecutionBudget(max_tokens=100)
        tracker = BudgetTracker(budget)
        tracker.record_tokens(150, "gpt-4o-mini")
        with self.assertRaises(BudgetExceededError) as ctx:
            tracker.check_tokens()
        self.assertEqual(ctx.exception.budget_type, "tokens")

    def test_time_limit(self):
        budget = ExecutionBudget(max_time_sec=0.01)
        tracker = BudgetTracker(budget)
        time.sleep(0.02)
        with self.assertRaises(BudgetExceededError) as ctx:
            tracker.check_time()
        self.assertEqual(ctx.exception.budget_type, "time")

    def test_cost_limit(self):
        budget = ExecutionBudget(max_cost_usd=0.001)
        tracker = BudgetTracker(budget)
        # gpt-4 is expensive: 0.03 per 1K tokens
        tracker.record_tokens(1000, "gpt-4")
        with self.assertRaises(BudgetExceededError) as ctx:
            tracker.check_cost()
        self.assertEqual(ctx.exception.budget_type, "cost")

    def test_check_all_records_and_checks(self):
        budget = ExecutionBudget(max_tokens=50)
        tracker = BudgetTracker(budget)
        tracker.record_step()
        with self.assertRaises(BudgetExceededError):
            tracker.check_all(new_tokens=100, model="gpt-4o-mini")

    def test_accessors(self):
        tracker = BudgetTracker(ExecutionBudget())
        tracker.record_step()
        tracker.record_step()
        tracker.record_tokens(500, "gpt-4o-mini")
        self.assertEqual(tracker.steps, 2)
        self.assertEqual(tracker.tokens, 500)
        self.assertGreater(tracker.cost, 0)
        self.assertGreater(tracker.elapsed_sec, 0)


if __name__ == "__main__":
    unittest.main()
