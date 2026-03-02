import hashlib
import logging
from collections import OrderedDict

from sentence_transformers import SentenceTransformer

logger = logging.getLogger(__name__)

MODEL_NAME = "sentence-transformers/all-MiniLM-L6-v2"
DEFAULT_CACHE_SIZE = 1024


class EmbeddingService:
    """Generates 384-dim embeddings using sentence-transformers with LRU cache."""

    def __init__(self, cache_size: int = DEFAULT_CACHE_SIZE):
        logger.info("Loading embedding model: %s", MODEL_NAME)
        self.model = SentenceTransformer(MODEL_NAME)
        logger.info("Embedding model loaded")
        self._cache: OrderedDict[str, list[float]] = OrderedDict()
        self._cache_size = cache_size
        self._hits = 0
        self._misses = 0

    @staticmethod
    def _cache_key(text: str) -> str:
        return hashlib.sha256(text.encode("utf-8")).hexdigest()[:16]

    @property
    def cache_stats(self) -> dict:
        return {
            "hits": self._hits,
            "misses": self._misses,
            "size": len(self._cache),
            "max_size": self._cache_size,
        }

    def embed(self, text: str) -> list[float]:
        """Generate a 384-dim embedding for a single text (cached)."""
        key = self._cache_key(text)
        if key in self._cache:
            self._hits += 1
            self._cache.move_to_end(key)
            return self._cache[key]

        self._misses += 1
        embedding = self.model.encode(text, normalize_embeddings=True).tolist()

        if len(self._cache) >= self._cache_size:
            self._cache.popitem(last=False)  # evict oldest
        self._cache[key] = embedding
        return embedding

    def embed_batch(self, texts: list[str]) -> list[list[float]]:
        """Generate embeddings for multiple texts."""
        results: list[list[float]] = []
        uncached_texts: list[str] = []
        uncached_indices: list[int] = []

        for i, text in enumerate(texts):
            key = self._cache_key(text)
            if key in self._cache:
                self._hits += 1
                self._cache.move_to_end(key)
                results.append(self._cache[key])
            else:
                self._misses += 1
                results.append([])  # placeholder
                uncached_texts.append(text)
                uncached_indices.append(i)

        if uncached_texts:
            embeddings = self.model.encode(uncached_texts, normalize_embeddings=True)
            for idx, embedding in zip(uncached_indices, embeddings):
                emb_list = embedding.tolist()
                key = self._cache_key(texts[idx])
                if len(self._cache) >= self._cache_size:
                    self._cache.popitem(last=False)
                self._cache[key] = emb_list
                results[idx] = emb_list

        return results
