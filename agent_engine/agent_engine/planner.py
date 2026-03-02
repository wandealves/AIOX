"""Lightweight task planner — decides whether to use tools, respond directly, or clarify."""

import json
import logging
from dataclasses import dataclass, field
from enum import Enum
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from .budget import BudgetTracker
    from .providers.base import LLMProvider

logger = logging.getLogger(__name__)


class ActionType(str, Enum):
    RESPOND = "respond"   # Direct response without tools
    EXECUTE = "execute"   # Use tools then respond
    CLARIFY = "clarify"   # Ask user for clarification


@dataclass
class PlannedAction:
    action: ActionType
    reasoning: str
    tools_to_use: list[str] = field(default_factory=list)
    response_hint: str = ""
    clarification: str = ""


_PLANNER_PROMPT = """\
You are a task planner. Given a user message and available tools, decide the best action.

Available tools: {tools}

Respond with ONLY a JSON object (no markdown fences):
{{
  "action": "respond" | "execute" | "clarify",
  "reasoning": "brief explanation",
  "tools_to_use": ["tool1", "tool2"],
  "response_hint": "optional hint for response",
  "clarification": "question to ask if action is clarify"
}}

Rules:
- "respond": Use when you can answer directly without tools.
- "execute": Use when tools are needed to answer properly.
- "clarify": Use when the request is too ambiguous to proceed.
"""


class Planner:
    """Decides how to handle a user request before the main LLM call."""

    def __init__(self, provider: "LLMProvider", planning_model: str = ""):
        self._provider = provider
        self._model = planning_model

    async def plan(
        self,
        user_message: str,
        system_prompt: str,
        available_tools: list[str],
        budget_tracker: "BudgetTracker | None" = None,
    ) -> PlannedAction:
        # Fast path: no tools → always RESPOND directly
        if not available_tools:
            return PlannedAction(
                action=ActionType.RESPOND,
                reasoning="No tools available",
            )

        tools_str = ", ".join(available_tools)
        planner_system = _PLANNER_PROMPT.format(tools=tools_str)

        try:
            response = await self._provider.generate(
                system_prompt=planner_system,
                user_message=user_message,
                model=self._model,
                temperature=0.0,
                max_tokens=512,
            )

            if response.error:
                logger.warning("Planner LLM error, defaulting to EXECUTE: %s", response.error)
                return PlannedAction(
                    action=ActionType.EXECUTE,
                    reasoning=f"Planner error: {response.error}",
                    tools_to_use=available_tools,
                )

            return _parse_plan(response.text, available_tools)

        except Exception as e:
            logger.warning("Planner failed, defaulting to EXECUTE: %s", e)
            return PlannedAction(
                action=ActionType.EXECUTE,
                reasoning=f"Planner exception: {e}",
                tools_to_use=available_tools,
            )


def _parse_plan(text: str, available_tools: list[str]) -> PlannedAction:
    """Parse a JSON response from the planner LLM."""
    # Try to extract JSON from possible markdown fences
    cleaned = text.strip()
    if cleaned.startswith("```"):
        # Remove ```json ... ``` wrapper
        lines = cleaned.split("\n")
        lines = [l for l in lines if not l.strip().startswith("```")]
        cleaned = "\n".join(lines).strip()

    try:
        data = json.loads(cleaned)
    except json.JSONDecodeError:
        logger.warning("Failed to parse planner JSON, defaulting to EXECUTE")
        return PlannedAction(
            action=ActionType.EXECUTE,
            reasoning="Failed to parse planner response",
            tools_to_use=available_tools,
        )

    action_str = data.get("action", "execute").lower()
    try:
        action = ActionType(action_str)
    except ValueError:
        action = ActionType.EXECUTE

    return PlannedAction(
        action=action,
        reasoning=data.get("reasoning", ""),
        tools_to_use=data.get("tools_to_use", []),
        response_hint=data.get("response_hint", ""),
        clarification=data.get("clarification", ""),
    )
