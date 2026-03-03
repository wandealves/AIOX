"""Tests for versioned tool registry."""

import unittest

from agent_engine.tools.base import BuiltinTool
from agent_engine.tools.registry import (
    _BUILTIN_TOOLS,
    get_all_tools,
    get_tool,
    is_builtin,
    register_tool,
)


class FakeTool(BuiltinTool):
    name = "fake_tool"
    description = "A fake tool"
    parameters = {"type": "object", "properties": {}}
    version = "1.0.0"

    async def execute(self, arguments, context=None):
        return {"status": "ok"}


class FakeToolV2(BuiltinTool):
    name = "fake_tool"
    description = "A fake tool v2"
    parameters = {"type": "object", "properties": {}}
    version = "2.0.0"

    async def execute(self, arguments, context=None):
        return {"status": "ok", "version": "2"}


class TestVersionedRegistry(unittest.TestCase):
    def setUp(self):
        # Save and clear registry for test isolation
        self._saved = {k: dict(v) for k, v in _BUILTIN_TOOLS.items()}
        _BUILTIN_TOOLS.clear()

    def tearDown(self):
        # Restore original registry state so other tests aren't affected
        _BUILTIN_TOOLS.clear()
        for k, v in self._saved.items():
            _BUILTIN_TOOLS[k] = v

    def test_register_and_get_by_name(self):
        register_tool(FakeTool())
        tool = get_tool("fake_tool")
        self.assertIsNotNone(tool)
        self.assertEqual(tool.name, "fake_tool")

    def test_get_by_version(self):
        register_tool(FakeTool())
        register_tool(FakeToolV2())
        v1 = get_tool("fake_tool", version="1.0.0")
        v2 = get_tool("fake_tool", version="2.0.0")
        self.assertEqual(v1.version, "1.0.0")
        self.assertEqual(v2.version, "2.0.0")

    def test_get_latest_version(self):
        register_tool(FakeTool())
        register_tool(FakeToolV2())
        latest = get_tool("fake_tool")
        self.assertEqual(latest.version, "2.0.0")

    def test_get_nonexistent(self):
        self.assertIsNone(get_tool("nonexistent"))

    def test_get_nonexistent_version(self):
        register_tool(FakeTool())
        self.assertIsNone(get_tool("fake_tool", version="9.9.9"))

    def test_is_builtin(self):
        register_tool(FakeTool())
        self.assertTrue(is_builtin("fake_tool"))
        self.assertTrue(is_builtin("fake-tool"))  # normalized
        self.assertFalse(is_builtin("nonexistent"))

    def test_get_all_tools_returns_latest(self):
        register_tool(FakeTool())
        register_tool(FakeToolV2())
        all_tools = get_all_tools()
        self.assertIn("fake_tool", all_tools)
        self.assertEqual(all_tools["fake_tool"].version, "2.0.0")

    def test_name_normalization(self):
        register_tool(FakeTool())
        self.assertIsNotNone(get_tool("Fake-Tool"))
        self.assertIsNotNone(get_tool("FAKE_TOOL"))


if __name__ == "__main__":
    unittest.main()
