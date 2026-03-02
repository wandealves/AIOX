"""Tests for model router and complexity estimator."""

import unittest

from agent_engine.model_router import ComplexityEstimator, route_model


class TestComplexityEstimator(unittest.TestCase):
    def test_simple_message(self):
        tier = ComplexityEstimator.estimate("Hi there!")
        self.assertEqual(tier, "small")

    def test_medium_with_tools_and_memory(self):
        tier = ComplexityEstimator.estimate(
            "Search for weather and compare results",
            has_tools=True,
            num_tools=2,
            has_memory=True,
        )
        self.assertEqual(tier, "medium")

    def test_large_complex_message(self):
        msg = "Analyze and compare the following data in detail. " + "x" * 2500
        tier = ComplexityEstimator.estimate(
            msg, has_tools=True, num_tools=5, has_memory=True
        )
        self.assertEqual(tier, "large")

    def test_keywords_boost_complexity(self):
        tier = ComplexityEstimator.estimate(
            "Please analyze in detail and create a comprehensive plan step by step",
        )
        self.assertIn(tier, ("medium", "large"))


class TestRouteModel(unittest.TestCase):
    def test_explicit_model_override(self):
        model = route_model("openai", "hi", explicit_model="custom-model")
        self.assertEqual(model, "custom-model")

    def test_simple_openai(self):
        model = route_model("openai", "Hello!")
        self.assertEqual(model, "gpt-4o-mini")

    def test_complex_anthropic(self):
        msg = "Analyze in detail and compare " + "x" * 3000
        model = route_model(
            "anthropic", msg, has_tools=True, num_tools=5, has_memory=True
        )
        self.assertEqual(model, "claude-opus-4-20250514")

    def test_unknown_provider_fallback(self):
        model = route_model("unknown_provider", "Hello!")
        # Falls back to openai tiers
        self.assertEqual(model, "gpt-4o-mini")


if __name__ == "__main__":
    unittest.main()
