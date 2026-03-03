"""Tests for Ollama text-based tool call parsing."""

import unittest

from agent_engine.providers.ollama import _parse_text_tool_calls, _try_parse_tool_call


SAMPLE_TOOLS = [
    {
        "type": "function",
        "function": {
            "name": "tavily_search",
            "description": "Search the web",
            "parameters": {"type": "object", "properties": {"query": {"type": "string"}}},
        },
    },
    {
        "type": "function",
        "function": {
            "name": "weather_api",
            "description": "Get weather",
            "parameters": {"type": "object", "properties": {"city": {"type": "string"}}},
        },
    },
]


class TestParseTextToolCalls(unittest.TestCase):
    def test_simple_json_with_name_and_parameters(self):
        text = '{"name": "tavily_search", "parameters": {"query": "capital da França"}}'
        result = _parse_text_tool_calls(text, SAMPLE_TOOLS)
        self.assertEqual(len(result), 1)
        self.assertEqual(result[0]["function"]["name"], "tavily_search")
        self.assertEqual(result[0]["function"]["arguments"]["query"], "capital da França")

    def test_json_with_prefix_text(self):
        """Model outputs text before the JSON (e.g. 'Ollama\\n{...}')."""
        text = 'Ollama\n{"name": "tavily-search", "parameters": {"query": "capital da França"}}'
        result = _parse_text_tool_calls(text, SAMPLE_TOOLS)
        self.assertEqual(len(result), 1)
        self.assertEqual(result[0]["function"]["name"], "tavily_search")

    def test_hyphenated_name_normalized(self):
        text = '{"name": "tavily-search", "parameters": {"query": "test"}}'
        result = _parse_text_tool_calls(text, SAMPLE_TOOLS)
        self.assertEqual(len(result), 1)
        self.assertEqual(result[0]["function"]["name"], "tavily_search")

    def test_arguments_key(self):
        """Model uses 'arguments' instead of 'parameters'."""
        text = '{"name": "tavily_search", "arguments": {"query": "test"}}'
        result = _parse_text_tool_calls(text, SAMPLE_TOOLS)
        self.assertEqual(len(result), 1)
        self.assertEqual(result[0]["function"]["arguments"]["query"], "test")

    def test_input_key(self):
        """Model uses 'input' instead of 'parameters'."""
        text = '{"name": "tavily_search", "input": {"query": "test"}}'
        result = _parse_text_tool_calls(text, SAMPLE_TOOLS)
        self.assertEqual(len(result), 1)
        self.assertEqual(result[0]["function"]["arguments"]["query"], "test")

    def test_unknown_tool_ignored(self):
        text = '{"name": "unknown_tool", "parameters": {"foo": "bar"}}'
        result = _parse_text_tool_calls(text, SAMPLE_TOOLS)
        self.assertEqual(len(result), 0)

    def test_no_json_returns_empty(self):
        text = "A capital da França é Paris."
        result = _parse_text_tool_calls(text, SAMPLE_TOOLS)
        self.assertEqual(len(result), 0)

    def test_empty_text(self):
        result = _parse_text_tool_calls("", SAMPLE_TOOLS)
        self.assertEqual(len(result), 0)

    def test_no_tools(self):
        text = '{"name": "tavily_search", "parameters": {"query": "test"}}'
        result = _parse_text_tool_calls(text, None)
        self.assertEqual(len(result), 0)

    def test_multiple_tool_calls(self):
        text = (
            'I will search for both. '
            '{"name": "tavily_search", "parameters": {"query": "weather"}} '
            '{"name": "weather_api", "parameters": {"city": "Paris"}}'
        )
        result = _parse_text_tool_calls(text, SAMPLE_TOOLS)
        self.assertEqual(len(result), 2)

    def test_json_with_no_name_field(self):
        text = '{"query": "test", "max_results": 5}'
        result = _parse_text_tool_calls(text, SAMPLE_TOOLS)
        self.assertEqual(len(result), 0)

    def test_nested_json_in_text(self):
        """JSON with nested objects should still be parsed correctly."""
        text = '{"name": "tavily_search", "parameters": {"query": "test", "options": {"depth": 1}}}'
        result = _parse_text_tool_calls(text, SAMPLE_TOOLS)
        self.assertEqual(len(result), 1)


class TestTryParseToolCall(unittest.TestCase):
    def test_valid(self):
        known = {"tavily_search"}
        result = _try_parse_tool_call('{"name": "tavily_search", "parameters": {"q": "hi"}}', known)
        self.assertIsNotNone(result)
        self.assertEqual(result["function"]["name"], "tavily_search")

    def test_invalid_json(self):
        result = _try_parse_tool_call("not json", {"tavily_search"})
        self.assertIsNone(result)

    def test_no_name(self):
        result = _try_parse_tool_call('{"parameters": {"q": "hi"}}', {"tavily_search"})
        self.assertIsNone(result)

    def test_unknown_name(self):
        result = _try_parse_tool_call('{"name": "nope", "parameters": {}}', {"tavily_search"})
        self.assertIsNone(result)


if __name__ == "__main__":
    unittest.main()
