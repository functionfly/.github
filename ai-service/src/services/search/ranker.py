"""Result ranker for semantic search.

Re-ranks search results by relevance using multiple factors.
"""

import logging
from typing import List, Dict, Any, Optional

logger = logging.getLogger(__name__)


class ResultRanker:
    """Re-ranks search results by relevance."""

    def __init__(self):
        # Weights for different ranking factors
        self.semantic_weight = 0.5
        self.keyword_weight = 0.3
        self.recency_weight = 0.2

    def rank_results(
        self,
        results: List[Dict[str, Any]],
        query: str,
        user_preferences: Optional[Dict[str, Any]] = None,
    ) -> List[Dict[str, Any]]:
        """Rank search results by relevance.

        Args:
            results: Initial search results from semantic search
            query: Original search query
            user_preferences: Optional user preferences for personalized ranking

        Returns:
            Re-ranked results with scores
        """
        if not results:
            return []

        query_terms = query.lower().split()

        # Calculate keyword match scores
        for result in results:
            data = result.get("data", {})
            keyword_score = self._calculate_keyword_score(data, query_terms)
            result["keyword_score"] = keyword_score

            # Calculate recency score
            recency_score = self._calculate_recency_score(data)
            result["recency_score"] = recency_score

            # Calculate final score
            final_score = (
                result["score"] * self.semantic_weight +
                keyword_score * self.keyword_weight +
                recency_score * self.recency_weight
            )
            result["final_score"] = final_score

        # Sort by final score
        results.sort(key=lambda x: x.get("final_score", 0), reverse=True)

        # Add rank position
        for i, result in enumerate(results):
            result["rank"] = i + 1

        return results

    def _calculate_keyword_score(
        self,
        data: Dict[str, Any],
        query_terms: List[str],
    ) -> float:
        """Calculate keyword match score.

        Args:
            data: Function data
            query_terms: Query terms

        Returns:
            Keyword match score (0-1)
        """
        if not query_terms:
            return 0.0

        matches = 0
        total_terms = len(query_terms)

        # Check name
        name = (data.get("name", "") or "").lower()
        for term in query_terms:
            if term in name:
                matches += 1
                continue
            # Check description
            description = (data.get("description", "") or "").lower()
            if term in description:
                matches += 0.5  # Lower weight for description matches
                continue
            # Check tags
            tags = data.get("tags", [])
            for tag in tags:
                if term in tag.lower():
                    matches += 0.8
                    break

        return min(matches / total_terms, 1.0)

    def _calculate_recency_score(self, data: Dict[str, Any]) -> float:
        """Calculate recency score based on last updated time.

        Args:
            data: Function data

        Returns:
            Recency score (0-1)
        """
        from datetime import datetime, timedelta

        try:
            indexed_at = data.get("indexed_at")
            if not indexed_at:
                return 0.5  # Default middle score

            # Parse the timestamp
            if isinstance(indexed_at, str):
                dt = datetime.fromisoformat(indexed_at.replace("Z", "+00:00"))
            else:
                dt = indexed_at

            # Calculate age in days
            age = (datetime.utcnow() - dt.replace(tzinfo=None)).days

            # Score decays over time - recent functions get higher scores
            if age <= 1:
                return 1.0
            elif age <= 7:
                return 0.8
            elif age <= 30:
                return 0.6
            elif age <= 90:
                return 0.4
            else:
                return 0.2

        except Exception as e:
            logger.warning(f"Failed to calculate recency: {e}")
            return 0.5

    def rerank_by_context(
        self,
        results: List[Dict[str, Any]],
        context: Dict[str, Any],
    ) -> List[Dict[str, Any]]:
        """Re-rank results based on contextual information.

        Args:
            results: Search results
            context: Context information (user history, current session, etc.)

        Returns:
            Re-ranked results
        """
        if not results or not context:
            return results

        # Get user's recently used functions
        recent_functions = context.get("recent_functions", [])
        if not recent_functions:
            return results

        # Boost recently used functions
        for result in results:
            function_id = result.get("function_id")
            if function_id in recent_functions:
                result["final_score"] = result.get("final_score", 0) * 1.2

        # Re-sort
        results.sort(key=lambda x: x.get("final_score", 0), reverse=True)

        # Update rank positions
        for i, result in enumerate(results):
            result["rank"] = i + 1

        return results

    def rank_results_triple(
        self,
        results: List[Dict[str, Any]],
        query: str,
        weights: Optional[Dict[str, float]] = None,
    ) -> List[Dict[str, Any]]:
        """Rank results using triple-vector scores + keyword + recency.

        Args:
            results: Triple search results with contract/semantic/code scores
            query: Original search query
            weights: Optional custom weights for contract/semantic/code

        Returns:
            Re-ranked results with final scores
        """
        if not results:
            return []

        if weights is None:
            weights = {"contract": 0.35, "semantic": 0.40, "code": 0.25}

        query_terms = query.lower().split()

        for result in results:
            # Triple score already computed in search_triple
            triple_score = result.get("triple_score", 0.0)

            # Keyword boost
            keyword_score = self._calculate_keyword_score(result.get("data", {}), query_terms)

            # Recency boost
            recency_score = self._calculate_recency_score(result.get("data", {}))

            result["final_score"] = (
                triple_score * 0.5
                + keyword_score * 0.3
                + recency_score * 0.2
            )
            result["keyword_score"] = keyword_score
            result["recency_score"] = recency_score

        results.sort(key=lambda x: x.get("final_score", 0), reverse=True)

        for i, result in enumerate(results):
            result["rank"] = i + 1

        return results

    def filter_results(
        self,
        results: List[Dict[str, Any]],
        filters: Dict[str, Any],
    ) -> List[Dict[str, Any]]:
        """Filter results based on criteria.

        Args:
            results: Search results
            filters: Filter criteria

        Returns:
            Filtered results
        """
        if not filters:
            return results

        filtered = []

        for result in results:
            data = result.get("data", {})

            # Runtime filter
            if runtime_filter := filters.get("runtime"):
                if data.get("runtime") != runtime_filter:
                    continue

            # Tags filter
            if tags_filter := filters.get("tags"):
                result_tags = set(data.get("tags", []))
                required_tags = set(tags_filter)
                if not result_tags.intersection(required_tags):
                    continue

            # Min score filter
            if min_score := filters.get("min_score"):
                if result.get("final_score", 0) < min_score:
                    continue

            filtered.append(result)

        return filtered


# Global instance
_result_ranker: Optional[ResultRanker] = None


def get_result_ranker() -> ResultRanker:
    """Get the global result ranker instance.

    Returns:
        The ResultRanker instance
    """
    global _result_ranker
    if _result_ranker is None:
        _result_ranker = ResultRanker()
    return _result_ranker