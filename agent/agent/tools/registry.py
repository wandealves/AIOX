"""Global registry for built-in tools."""

from __future__ import annotations

import logging
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from .base import BuiltinTool

logger = logging.getLogger(__name__)

_BUILTIN_TOOLS: dict[str, BuiltinTool] = {}


def _normalize(name: str) -> str:
    """Normalize tool name: hyphens → underscores, lowercase."""
    return name.replace("-", "_").lower()


def register_tool(tool: BuiltinTool) -> None:
    """Register a built-in tool by its normalized name."""
    key = _normalize(tool.name)
    if key in _BUILTIN_TOOLS:
        logger.warning("Overwriting existing built-in tool: %s", key)
    _BUILTIN_TOOLS[key] = tool
    logger.debug("Registered built-in tool: %s", key)


def get_tool(name: str) -> BuiltinTool | None:
    """Return a built-in tool by name, or None if not registered.

    Normalizes the name so both ``tavily-search`` and ``tavily_search`` match.
    """
    return _BUILTIN_TOOLS.get(_normalize(name))


def is_builtin(name: str) -> bool:
    """Check whether a tool name corresponds to a registered built-in tool."""
    return _normalize(name) in _BUILTIN_TOOLS


def get_all_tools() -> dict[str, BuiltinTool]:
    """Return a copy of all registered built-in tools."""
    return dict(_BUILTIN_TOOLS)
