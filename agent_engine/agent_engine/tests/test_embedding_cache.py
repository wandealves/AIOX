"""Tests for embedding cache (LRU)."""

import unittest
from unittest.mock import MagicMock, patch

from agent_engine.embedding import EmbeddingService


class TestEmbeddingCache(unittest.TestCase):
    @patch("agent_engine.embedding.SentenceTransformer")
    def setUp(self, mock_st_cls):
        """Set up with mocked SentenceTransformer to avoid loading the real model."""
        import numpy as np

        self.mock_model = MagicMock()
        self.mock_model.encode.side_effect = lambda text, **kw: (
            np.random.rand(384).astype("float32")
            if isinstance(text, str)
            else np.random.rand(len(text), 384).astype("float32")
        )
        mock_st_cls.return_value = self.mock_model
        self.svc = EmbeddingService(cache_size=3)

    def test_cache_miss(self):
        result = self.svc.embed("hello")
        self.assertEqual(len(result), 384)
        self.assertEqual(self.svc.cache_stats["misses"], 1)
        self.assertEqual(self.svc.cache_stats["hits"], 0)

    def test_cache_hit(self):
        r1 = self.svc.embed("hello")
        r2 = self.svc.embed("hello")
        self.assertEqual(r1, r2)
        self.assertEqual(self.svc.cache_stats["hits"], 1)
        self.assertEqual(self.svc.cache_stats["misses"], 1)
        # Model should have been called only once
        self.assertEqual(self.mock_model.encode.call_count, 1)

    def test_lru_eviction(self):
        self.svc.embed("a")
        self.svc.embed("b")
        self.svc.embed("c")
        self.assertEqual(self.svc.cache_stats["size"], 3)

        # Adding a 4th should evict "a"
        self.svc.embed("d")
        self.assertEqual(self.svc.cache_stats["size"], 3)

        # "a" should now be a miss
        self.svc.embed("a")
        self.assertEqual(self.svc.cache_stats["misses"], 5)  # a,b,c,d,a

    def test_cache_stats(self):
        stats = self.svc.cache_stats
        self.assertEqual(stats["size"], 0)
        self.assertEqual(stats["max_size"], 3)
        self.assertEqual(stats["hits"], 0)
        self.assertEqual(stats["misses"], 0)


if __name__ == "__main__":
    unittest.main()
