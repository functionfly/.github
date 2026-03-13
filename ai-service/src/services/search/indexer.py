"""Search indexer for semantic function search.

Builds and maintains the search index for function registry using embeddings.
"""

import json
import logging
from typing import Optional, Any, Dict, List
from datetime import datetime

import redis.asyncio as redis

from ...config import settings
from ...integrations.orchestrator.client import get_orchestrator_client
from ..embeddings import get_embeddings_service

logger = logging.getLogger(__name__)

# Redis key prefixes
INDEX_KEY_PREFIX = "search:index:"
FUNCTION_DATA_PREFIX = "search:function:"
USER_INDEX_PREFIX = "search:user:"


class SearchIndexer:
    """Builds and maintains semantic search index."""

    def __init__(self):
        self._redis: Optional[redis.Redis] = None
        self._orchestrator = get_orchestrator_client()
        self._embeddings_service = get_embeddings_service()

    async def get_redis(self) -> Optional[redis.Redis]:
        """Get Redis connection."""
        if self._redis is None:
            try:
                self._redis = redis.from_url(
                    settings.redis_url,
                    encoding="utf-8",
                    decode_responses=True,
                )
                await self._redis.ping()
            except Exception as e:
                logger.warning(f"Failed to connect to Redis: {e}")
                self._redis = None
        return self._redis

    def _get_index_key(self, tenant_id: str) -> str:
        """Get Redis key for tenant's search index."""
        return f"{INDEX_KEY_PREFIX}{tenant_id}"

    def _get_function_key(self, function_id: str) -> str:
        """Get Redis key for function data."""
        return f"{FUNCTION_DATA_PREFIX}{function_id}"

    def _get_user_index_key(self, tenant_id: str) -> str:
        """Get Redis key for user's function index."""
        return f"{USER_INDEX_PREFIX}{tenant_id}"

    async def index_function(
        self,
        function_id: str,
        function_data: Dict[str, Any],
    ) -> bool:
        """Index a single function.

        Args:
            function_id: The function ID
            function_data: Function data including name, description, etc.

        Returns:
            True if indexed successfully
        """
        try:
            # Create searchable text from function data
            searchable_text = self._create_searchable_text(function_data)

            # Generate embedding
            from ...models.schemas import EmbeddingRequest
            request = EmbeddingRequest(text=searchable_text)
            embedding_response = await self._embeddings_service.generate_embedding(request)
            embedding = embedding_response.embedding

            # Store in Redis
            redis_client = await self.get_redis()
            if not redis_client:
                logger.error("Redis not available for indexing")
                return False

            # Store function data
            function_storage = {
                "function_id": function_id,
                "name": function_data.get("name", ""),
                "description": function_data.get("description", ""),
                "runtime": function_data.get("runtime", ""),
                "tags": function_data.get("tags", []),
                "metadata": function_data.get("metadata", {}),
                "indexed_at": datetime.utcnow().isoformat(),
            }
            await redis_client.hset(
                self._get_function_key(function_id),
                mapping={
                    "data": json.dumps(function_storage),
                    "embedding": json.dumps(embedding),
                },
            )

            logger.info(f"Indexed function {function_id}")
            return True

        except Exception as e:
            logger.error(f"Failed to index function {function_id}: {e}")
            return False

    def _create_searchable_text(self, function_data: Dict[str, Any]) -> str:
        """Create searchable text from function data.

        Args:
            function_data: Function data

        Returns:
            Searchable text string
        """
        parts = []

        # Add name
        if name := function_data.get("name"):
            parts.append(f"Function name: {name}")

        # Add description
        if desc := function_data.get("description"):
            parts.append(f"Description: {desc}")

        # Add runtime
        if runtime := function_data.get("runtime"):
            parts.append(f"Runtime: {runtime}")

        # Add tags
        if tags := function_data.get("tags"):
            parts.append(f"Tags: {', '.join(tags)}")

        # Add metadata
        if metadata := function_data.get("metadata"):
            for key, value in metadata.items():
                parts.append(f"{key}: {value}")

        return " ".join(parts)

    async def index_tenant_functions(
        self,
        tenant_id: str,
    ) -> int:
        """Index all functions for a tenant.

        Args:
            tenant_id: The tenant ID

        Returns:
            Number of functions indexed
        """
        try:
            # Get functions from orchestrator
            functions = await self._orchestrator.get_functions_by_tenant(
                tenant_id,
                limit=100,
            )

            if not functions:
                logger.info(f"No functions found for tenant {tenant_id}")
                return 0

            # Index each function
            indexed_count = 0
            for func in functions:
                function_id = func.get("id")
                if function_id:
                    success = await self.index_function(function_id, func)
                    if success:
                        indexed_count += 1

            # Update tenant's function list in Redis
            redis_client = await self.get_redis()
            if redis_client:
                function_ids = [f.get("id") for f in functions if f.get("id")]
                await redis_client.sadd(
                    self._get_user_index_key(tenant_id),
                    *function_ids,
                )

            logger.info(f"Indexed {indexed_count} functions for tenant {tenant_id}")
            return indexed_count

        except Exception as e:
            logger.error(f"Failed to index tenant functions: {e}")
            return 0

    async def search(
        self,
        tenant_id: str,
        query_embedding: List[float],
        limit: int = 20,
    ) -> List[Dict[str, Any]]:
        """Search functions by semantic similarity.

        Args:
            tenant_id: The tenant ID
            query_embedding: Query embedding vector
            limit: Maximum results to return

        Returns:
            List of matching functions with scores
        """
        redis_client = await self.get_redis()
        if not redis_client:
            logger.error("Redis not available for search")
            return []

        try:
            # Get all function IDs for this tenant
            function_ids = await redis_client.smembers(
                self._get_user_index_key(tenant_id),
            )

            if not function_ids:
                return []

            results = []

            # Compare with each function's embedding
            for function_id in function_ids:
                func_data = await redis_client.hgetall(
                    self._get_function_key(function_id),
                )

                if not func_data or "embedding" not in func_data:
                    continue

                try:
                    stored_embedding = json.loads(func_data["embedding"])
                    function_data = json.loads(func_data["data"])

                    # Calculate cosine similarity
                    similarity = self._cosine_similarity(query_embedding, stored_embedding)

                    results.append({
                        "function_id": function_id,
                        "data": function_data,
                        "score": similarity,
                    })
                except Exception as e:
                    logger.warning(f"Failed to process function {function_id}: {e}")
                    continue

            # Sort by score descending
            results.sort(key=lambda x: x["score"], reverse=True)

            return results[:limit]

        except Exception as e:
            logger.error(f"Search failed: {e}")
            return []

    def _cosine_similarity(
        self,
        vec1: List[float],
        vec2: List[float],
    ) -> float:
        """Calculate cosine similarity between two vectors.

        Args:
            vec1: First vector
            vec2: Second vector

        Returns:
            Cosine similarity score
        """
        if not vec1 or not vec2 or len(vec1) != len(vec2):
            return 0.0

        dot_product = sum(a * b for a, b in zip(vec1, vec2))
        magnitude1 = sum(a * a for a in vec1) ** 0.5
        magnitude2 = sum(b * b for b in vec2) ** 0.5

        if magnitude1 == 0 or magnitude2 == 0:
            return 0.0

        return dot_product / (magnitude1 * magnitude2)

    async def remove_function(self, function_id: str, tenant_id: str) -> bool:
        """Remove a function from the index.

        Args:
            function_id: The function ID
            tenant_id: The tenant ID

        Returns:
            True if removed successfully
        """
        redis_client = await self.get_redis()
        if not redis_client:
            return False

        try:
            # Delete function data
            await redis_client.delete(self._get_function_key(function_id))
            # Remove from tenant's function set
            await redis_client.srem(
                self._get_user_index_key(tenant_id),
                function_id,
            )
            logger.info(f"Removed function {function_id} from index")
            return True
        except Exception as e:
            logger.error(f"Failed to remove function: {e}")
            return False

    async def get_index_stats(self, tenant_id: str) -> Dict[str, Any]:
        """Get index statistics for a tenant.

        Args:
            tenant_id: The tenant ID

        Returns:
            Index statistics
        """
        redis_client = await self.get_redis()
        if not redis_client:
            return {"error": "Redis not available"}

        try:
            function_ids = await redis_client.smembers(
                self._get_user_index_key(tenant_id),
            )
            return {
                "tenant_id": tenant_id,
                "function_count": len(function_ids),
                "indexed_at": datetime.utcnow().isoformat(),
            }
        except Exception as e:
            logger.error(f"Failed to get index stats: {e}")
            return {"error": str(e)}

    async def close(self):
        """Close Redis connection."""
        if self._redis:
            await self._redis.close()
            self._redis = None


# Global instance
_search_indexer: Optional[SearchIndexer] = None


def get_search_indexer() -> SearchIndexer:
    """Get the global search indexer instance.

    Returns:
        The SearchIndexer instance
    """
    global _search_indexer
    if _search_indexer is None:
        _search_indexer = SearchIndexer()
    return _search_indexer
