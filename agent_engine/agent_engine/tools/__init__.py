"""Tools package — built-in tool registry and HTTP/built-in executor."""

from .base import BuiltinTool
from .executor import (
    ToolExecutor,
    proto_tools_to_anthropic_format,
    proto_tools_to_openai_format,
)
from .registry import get_all_tools, get_tool, is_builtin, register_tool

# ---------------------------------------------------------------------------
# Auto-register all built-in tools on import
# ---------------------------------------------------------------------------
from .tavily import TavilySearchTool

register_tool(TavilySearchTool())

__all__ = [
    "BuiltinTool",
    "ToolExecutor",
    "proto_tools_to_anthropic_format",
    "proto_tools_to_openai_format",
    "register_tool",
    "get_tool",
    "is_builtin",
    "get_all_tools",
]
