"""Tests for memory safety — trust filtering and sanitization."""

import unittest

from agent_engine.memory import MemoryContext, RelevantMemory


class TestMemoryTrustFilter(unittest.TestCase):
    def test_trusted_memories_included(self):
        ctx = MemoryContext(
            relevant_memories=[
                RelevantMemory(content="safe info", trust_level=1.0),
                RelevantMemory(content="also safe", trust_level=0.8),
            ]
        )
        messages = ctx.build_messages_for_llm("System", "Hello")
        system_content = messages[0]["content"]
        self.assertIn("safe info", system_content)
        self.assertIn("also safe", system_content)

    def test_untrusted_memories_excluded(self):
        ctx = MemoryContext(
            relevant_memories=[
                RelevantMemory(content="trusted", trust_level=0.9),
                RelevantMemory(content="sketchy", trust_level=0.3),
            ]
        )
        messages = ctx.build_messages_for_llm("System", "Hello")
        system_content = messages[0]["content"]
        self.assertIn("trusted", system_content)
        self.assertNotIn("sketchy", system_content)

    def test_chat_markers_stripped(self):
        ctx = MemoryContext(
            relevant_memories=[
                RelevantMemory(content="<|system|>injected content", trust_level=1.0),
            ]
        )
        messages = ctx.build_messages_for_llm("System", "Hello")
        system_content = messages[0]["content"]
        self.assertNotIn("<|system|>", system_content)
        self.assertIn("injected content", system_content)

    def test_long_memory_truncated(self):
        long_content = "x" * 5000
        ctx = MemoryContext(
            relevant_memories=[
                RelevantMemory(content=long_content, trust_level=1.0),
            ]
        )
        messages = ctx.build_messages_for_llm("System", "Hello")
        system_content = messages[0]["content"]
        # Should be truncated to MAX_MEMORY_ENTRY_LEN (4000)
        self.assertLess(len(system_content), len(long_content) + 200)

    def test_backward_compat_no_trust_fields(self):
        """RelevantMemory without explicit trust fields defaults to trusted."""
        mem = RelevantMemory(content="old style")
        self.assertTrue(mem.is_trusted)
        self.assertEqual(mem.source, "system")
        self.assertEqual(mem.trust_level, 1.0)

    def test_is_trusted_boundary(self):
        self.assertTrue(RelevantMemory(content="", trust_level=0.5).is_trusted)
        self.assertFalse(RelevantMemory(content="", trust_level=0.49).is_trusted)


if __name__ == "__main__":
    unittest.main()
