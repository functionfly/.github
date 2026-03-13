"""Query processor for semantic search.

Processes natural language queries for function search.
"""

import logging
import re
from typing import Optional, List, Dict, Any, Tuple

from ..embeddings import get_embeddings_service

logger = logging.getLogger(__name__)


class QueryProcessor:
    """Processes natural language queries for semantic search."""

    def __init__(self):
        self._embeddings_service = get_embeddings_service()

    async def process_query(
        self,
        query: str,
        user_id: Optional[str] = None,
    ) -> Tuple[str, List[float], Dict[str, Any]]:
        """Process a natural language query.

        Args:
            query: Natural language search query
            user_id: Optional user ID for personalized results

        Returns:
            Tuple of (processed_query, embedding, metadata)
        """
        # Clean and normalize the query
        processed_query = self._clean_query(query)

        # Expand query with synonyms
        expanded_query = self._expand_query(processed_query)

        # Generate embedding
        from ...models.schemas import EmbeddingRequest
        request = EmbeddingRequest(text=expanded_query)
        embedding_response = await self._embeddings_service.generate_embedding(request)
        embedding = embedding_response.embedding

        # Extract metadata
        metadata = self._extract_query_metadata(query)

        if user_id:
            metadata["user_id"] = user_id

        return processed_query, embedding, metadata

    def _clean_query(self, query: str) -> str:
        """Clean and normalize the query.

        Args:
            query: Raw query string

        Returns:
            Cleaned query
        """
        # Remove extra whitespace
        query = " ".join(query.split())

        # Remove special characters that might interfere
        query = re.sub(r'[^\w\s\-_.]', '', query)

        # Convert to lowercase for matching
        return query.strip()

    def _expand_query(self, query: str) -> str:
        """Expand query with synonyms and related terms.

        Args:
            query: Cleaned query

        Returns:
            Expanded query string
        """
        # Common synonym mappings for function search
        synonyms = {
            "email": ["email", "send email", "notification", "mail"],
            "send": ["send", "dispatch", "deliver", "notify"],
            "http": ["http", "api", "request", "webhook", "rest"],
            "process": ["process", "handle", "execute", "run"],
            "data": ["data", "database", "storage", "record"],
            "transform": ["transform", "convert", "map", "parse"],
            "validate": ["validate", "check", "verify", "sanitize"],
            "auth": ["auth", "authenticate", "login", "security"],
            "image": ["image", "picture", "photo", "thumbnail"],
            "file": ["file", "upload", "download", "storage"],
            "queue": ["queue", "async", "background", "job"],
            "schedule": ["schedule", "cron", "timer", "periodic"],
        }

        query_lower = query.lower()
        expansion_parts = [query]

        # Add synonyms for matched terms
        for key, values in synonyms.items():
            if key in query_lower:
                expansion_parts.extend(values)

        return " ".join(expansion_parts)

    def _extract_query_metadata(self, query: str) -> Dict[str, Any]:
        """Extract metadata from query.

        Args:
            query: Original query

        Returns:
            Metadata dictionary
        """
        metadata: Dict[str, Any] = {}

        # Detect runtime hints
        runtimes = ["python", "node", "javascript", "go", "rust", "ruby"]
        for runtime in runtimes:
            if runtime in query.lower():
                metadata["runtime"] = runtime
                break

        # Detect intent keywords
        intent_keywords = {
            "api": "api",
            "webhook": "webhook",
            "cron": "scheduled",
            "background": "async",
            "transform": "transform",
            "process": "processing",
        }
        for key, intent in intent_keywords.items():
            if key in query.lower():
                metadata.setdefault("intents", []).append(intent)

        # Detect if it's a capability search vs name search
        # Capability search: "functions that send email"
        # Name search: "email-sender-function"
        capability_patterns = [
            r"functions? (that|which|to)",
            r"can (you|i) ",
            r"send\s+\w+",
            r"process\s+\w+",
            r"handle\s+\w+",
        ]
        for pattern in capability_patterns:
            if re.search(pattern, query.lower()):
                metadata["search_type"] = "capability"
                break
        else:
            metadata["search_type"] = "name"

        return metadata

    def process_keyword_fallback(
        self,
        query: str,
        functions: List[Dict[str, Any]],
    ) -> List[Dict[str, Any]]:
        """Fallback to keyword search when semantic search fails.

        Args:
            query: Search query
            functions: List of functions to filter

        Returns:
            Filtered functions with scores
        """
        query_terms = query.lower().split()
        results = []

        for func in functions:
            data = func.get("data", {})
            score = 0.0

            # Check name
            name = (data.get("name", "") or "").lower()
            for term in query_terms:
                if term in name:
                    score += 1.0

            # Check description
            desc = (data.get("description", "") or "").lower()
            for term in query_terms:
                if term in desc:
                    score += 0.5

            # Check tags
            tags = data.get("tags", [])
            for tag in tags:
                tag_lower = tag.lower()
                for term in query_terms:
                    if term in tag_lower:
                        score += 0.8

            if score > 0:
                results.append({
                    **func,
                    "score": score,
                    "match_type": "keyword",
                })

        # Sort by score
        results.sort(key=lambda x: x.get("score", 0), reverse=True)

        return results[:20]  # Limit to top 20

    def suggest_corrections(self, query: str) -> List[str]:
        """Suggest query corrections/alternatives.

        Args:
            query: Original query

        Returns:
            List of suggested alternative queries
        """
        suggestions = []

        # Common typo corrections
        corrections = {
            "functoin": "function",
            "functon": "function",
            "funtion": "function",
            "functiion": "function",
            "deploy": "deploy",
            "deploye": "deploy",
            "functoin": "function",
        }

        for wrong, correct in corrections.items():
            if wrong in query.lower():
                suggestion = query.lower().replace(wrong, correct)
                suggestions.append(suggestion)
                break

        return suggestions


# Global instance
_query_processor: Optional[QueryProcessor] = None


def get_query_processor() -> QueryProcessor:
    """Get the global query processor instance.

    Returns:
        The QueryProcessor instance
    """
    global _query_processor
    if _query_processor is None:
        _query_processor = QueryProcessor()
    return _query_processor
