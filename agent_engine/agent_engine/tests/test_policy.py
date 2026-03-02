"""Tests for the policy engine."""

import unittest

from agent_engine.policy import Policy


class TestPolicy(unittest.TestCase):
    def test_from_permissions_none(self):
        p = Policy.from_permissions(None)
        self.assertFalse(p.blocked)
        self.assertEqual(p.allowed_providers, [])
        self.assertEqual(p.allowed_domains, [])

    def test_from_permissions_full(self):
        p = Policy.from_permissions(
            {
                "allowed_providers": ["openai", "anthropic"],
                "allowed_domains": ["example.com"],
                "blocked": True,
            }
        )
        self.assertTrue(p.blocked)
        self.assertEqual(p.allowed_providers, ["openai", "anthropic"])

    def test_check_provider_empty_allows_all(self):
        p = Policy()
        self.assertTrue(p.check_provider("anything"))

    def test_check_provider_match(self):
        p = Policy(allowed_providers=["openai", "anthropic"])
        self.assertTrue(p.check_provider("openai"))
        self.assertFalse(p.check_provider("gemini"))

    def test_check_tool_url_empty_allows_all(self):
        p = Policy()
        self.assertTrue(p.check_tool_url("https://evil.com/api"))

    def test_check_tool_url_exact_match(self):
        p = Policy(allowed_domains=["example.com"])
        self.assertTrue(p.check_tool_url("https://example.com/api"))
        self.assertFalse(p.check_tool_url("https://evil.com/api"))

    def test_check_tool_url_subdomain_match(self):
        p = Policy(allowed_domains=["example.com"])
        self.assertTrue(p.check_tool_url("https://api.example.com/v1"))
        self.assertTrue(p.check_tool_url("https://deep.sub.example.com/v1"))

    def test_check_tool_url_no_false_positive(self):
        p = Policy(allowed_domains=["example.com"])
        # notexample.com should NOT match example.com
        self.assertFalse(p.check_tool_url("https://notexample.com/api"))

    def test_check_tool_url_invalid(self):
        p = Policy(allowed_domains=["example.com"])
        self.assertFalse(p.check_tool_url("not-a-url"))


if __name__ == "__main__":
    unittest.main()
