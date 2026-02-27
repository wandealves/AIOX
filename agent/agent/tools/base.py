"""Base class for all built-in tools."""

from abc import ABC, abstractmethod


class BuiltinTool(ABC):
    """Abstract base class for built-in tools.

    Subclasses must define class attributes `name`, `description`,
    and `parameters`, and implement the async `execute` method.
    """

    name: str
    description: str
    parameters: dict  # JSON Schema for the tool arguments

    @abstractmethod
    async def execute(self, arguments: dict, context: dict | None = None) -> dict:
        """Execute the tool with the given arguments.

        Returns a dict with keys: status, data, error.
        """
        ...
