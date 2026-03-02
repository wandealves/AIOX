"""Global registry for built-in tools with version support."""

from __future__ import annotations

import logging
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from .base import BuiltinTool

logger = logging.getLogger(__name__)

# name -> version -> tool
_BUILTIN_TOOLS: dict[str, dict[str, BuiltinTool]] = {}


def _normalize(name: str) -> str:
    """Normalize tool name: hyphens -> underscores, lowercase."""
    return name.replace("-", "_").lower()


def register_tool(tool: BuiltinTool) -> None:
    """Register a built-in tool by its normalized name and version."""
    key = _normalize(tool.name)
    version = getattr(tool, "version", "1.0.0")
    if key not in _BUILTIN_TOOLS:
        _BUILTIN_TOOLS[key] = {}
    if version in _BUILTIN_TOOLS[key]:
        logger.warning("Overwriting built-in tool: %s@%s", key, version)
    _BUILTIN_TOOLS[key][version] = tool
    logger.debug("Registered built-in tool: %s@%s", key, version)


def get_tool(name: str, version: str | None = None) -> BuiltinTool | None:
    """Return a built-in tool by name, or None if not registered.

    If *version* is None, returns the latest version (highest semver string).
    Normalizes the name so both ``tavily-search`` and ``tavily_search`` match.
    """
    versions = _BUILTIN_TOOLS.get(_normalize(name))
    if not versions:
        return None
    if version is not None:
        return versions.get(version)
    # Return latest version (simple string sort works for semver with same digit count)
    latest = max(versions.keys())
    return versions[latest]


def is_builtin(name: str) -> bool:
    """Check whether a tool name corresponds to a registered built-in tool."""
    return _normalize(name) in _BUILTIN_TOOLS


def get_all_tools() -> dict[str, BuiltinTool]:
    """Return a flat dict of all registered built-in tools (latest versions).

    Keyed by normalized name for backward compatibility.
    """
    result: dict[str, BuiltinTool] = {}
    for name, versions in _BUILTIN_TOOLS.items():
        latest = max(versions.keys())
        result[name] = versions[latest]
    return result
