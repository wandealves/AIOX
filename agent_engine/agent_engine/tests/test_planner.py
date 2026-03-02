"""Tests for the task planner."""

import asyncio
import json
import unittest
from unittest.mock import AsyncMock, MagicMock

from agent_engine.planner import ActionType, PlannedAction, Planner, _parse_plan
from agent_engine.providers.base import LLMResponse


class TestParsePlan(unittest.TestCase):
    def test_valid_json(self):
        text = json.dumps(
            {
                "action": "respond",
                "reasoning": "Simple question",
                "tools_to_use": [],
            }
        )
        result = _parse_plan(text, ["tavily_search"])
        self.assertEqual(result.action, ActionType.RESPOND)
        self.assertEqual(result.reasoning, "Simple question")

    def test_markdown_fenced_json(self):
        text = '```json\n{"action": "execute", "reasoning": "need search", "tools_to_use": ["tavily_search"]}\n```'
        result = _parse_plan(text, ["tavily_search"])
        self.assertEqual(result.action, ActionType.EXECUTE)
        self.assertIn("tavily_search", result.tools_to_use)

    def test_invalid_json_falls_back_to_execute(self):
        result = _parse_plan("this is not json", ["tool1"])
        self.assertEqual(result.action, ActionType.EXECUTE)
        self.assertEqual(result.tools_to_use, ["tool1"])

    def test_clarify_action(self):
        text = json.dumps(
            {
                "action": "clarify",
                "reasoning": "ambiguous",
                "clarification": "What do you mean?",
            }
        )
        result = _parse_plan(text, [])
        self.assertEqual(result.action, ActionType.CLARIFY)
        self.assertEqual(result.clarification, "What do you mean?")

    def test_unknown_action_defaults_to_execute(self):
        text = json.dumps({"action": "unknown_action", "reasoning": "test"})
        result = _parse_plan(text, ["t1"])
        self.assertEqual(result.action, ActionType.EXECUTE)


class TestPlanner(unittest.TestCase):
    def test_no_tools_returns_respond(self):
        """Without tools, planner skips LLM call and returns RESPOND."""
        provider = MagicMock()
        planner = Planner(provider=provider)

        result = asyncio.get_event_loop().run_until_complete(
            planner.plan("Hello", "You are helpful", [], None)
        )
        self.assertEqual(result.action, ActionType.RESPOND)
        provider.generate.assert_not_called()

    def test_with_tools_calls_provider(self):
        """With tools, planner calls provider and parses the result."""
        provider = MagicMock()
        provider.generate = AsyncMock(
            return_value=LLMResponse(
                text=json.dumps(
                    {
                        "action": "execute",
                        "reasoning": "need search",
                        "tools_to_use": ["search"],
                    }
                ),
                tokens_used=50,
                model_used="test",
                duration_ms=100,
            )
        )
        planner = Planner(provider=provider)

        result = asyncio.get_event_loop().run_until_complete(
            planner.plan("Search for cats", "system", ["search"], None)
        )
        self.assertEqual(result.action, ActionType.EXECUTE)
        provider.generate.assert_called_once()

    def test_provider_error_defaults_to_execute(self):
        """If provider returns an error, default to EXECUTE."""
        provider = MagicMock()
        provider.generate = AsyncMock(
            return_value=LLMResponse(
                text="",
                tokens_used=0,
                model_used="test",
                duration_ms=0,
                error="API error",
            )
        )
        planner = Planner(provider=provider)

        result = asyncio.get_event_loop().run_until_complete(
            planner.plan("msg", "sys", ["tool1"], None)
        )
        self.assertEqual(result.action, ActionType.EXECUTE)

    def test_provider_exception_defaults_to_execute(self):
        """If provider raises, default to EXECUTE."""
        provider = MagicMock()
        provider.generate = AsyncMock(side_effect=RuntimeError("boom"))
        planner = Planner(provider=provider)

        result = asyncio.get_event_loop().run_until_complete(
            planner.plan("msg", "sys", ["tool1"], None)
        )
        self.assertEqual(result.action, ActionType.EXECUTE)


if __name__ == "__main__":
    unittest.main()
