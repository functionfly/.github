"""Recommendation engine — ALS collaborative filtering + content-based fallback.

Provides personalized function recommendations based on user interaction history.
Falls back to content-based (FlyEmbed similarity) for cold-start users.
"""

import json
import logging
from collections import defaultdict
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Tuple

import numpy as np
import redis.asyncio as redis
from scipy.sparse import csr_matrix

from ...config import settings
from .models import InteractionEvent, RecommendationResult, RecommendationResponse

logger = logging.getLogger(__name__)


class RecommendationEngine:
    """ALS-based collaborative filtering with content-based fallback.

    Maintains a user-function interaction matrix in Redis.
    Trains ALS factors periodically and serves recommendations.
    Falls back to popularity/content-based for cold-start users.
    """

    INTERACTIONS_KEY = "ml:recommendations:interactions"
    FACTORS_USER_KEY = "ml:recommendations:users"
    FACTORS_ITEM_KEY = "ml:recommendations:items"
    POPULARITY_KEY = "ml:recommendations:popularity"
    USER_MAP_KEY = "ml:recommendations:user_map"
    ITEM_MAP_KEY = "ml:recommendations:item_map"
    EXPIRY_DAYS = 90

    # Interaction weights
    INTERACTION_WEIGHTS = {
        "install": 5.0,
        "execute": 3.0,
        "rate": 4.0,
        "search_click": 2.0,
        "view": 1.0,
        "search_impression": 0.5,
    }

    def __init__(self):
        self._redis: Optional[redis.Redis] = None
        self._latent_dims = settings.ml_recommendation_latent_dims
        self._user_factors: Optional[np.ndarray] = None
        self._item_factors: Optional[np.ndarray] = None
        self._user_map: Dict[str, int] = {}
        self._item_map: Dict[str, int] = {}
        self._reverse_item_map: Dict[int, str] = {}

    async def get_redis(self) -> Optional[redis.Redis]:
        if self._redis is None:
            try:
                self._redis = redis.from_url(
                    settings.redis_url, encoding="utf-8", decode_responses=True
                )
                await self._redis.ping()
            except Exception as e:
                logger.warning(f"Redis connection failed: {e}")
                self._redis = None
        return self._redis

    async def record_interaction(self, event: InteractionEvent) -> None:
        """Record a user-function interaction."""
        r = await self.get_redis()
        if not r:
            return

        try:
            weight = self.INTERACTION_WEIGHTS.get(event.interaction_type, 1.0)
            entry = json.dumps({
                "user_id": event.user_id,
                "function_id": event.function_id,
                "type": event.interaction_type,
                "weight": weight,
                "ts": event.timestamp.isoformat(),
            })
            score = event.timestamp.timestamp()
            await r.zadd(self.INTERACTIONS_KEY, {entry: score})

            # Update popularity counter
            await r.zincrby(self.POPULARITY_KEY, weight, event.function_id)

            # Invalidate cached factors
            await r.delete(self.FACTORS_USER_KEY, self.FACTORS_ITEM_KEY)
        except Exception as e:
            logger.error(f"Failed to record interaction: {e}")

    async def _build_interaction_matrix(self) -> Tuple[csr_matrix, Dict[str, int], Dict[str, int]]:
        """Build user-function interaction matrix from Redis."""
        r = await self.get_redis()
        if not r:
            return csr_matrix((0, 0)), {}, {}

        try:
            items = await r.zrange(self.INTERACTIONS_KEY, 0, -1)

            user_map: Dict[str, int] = {}
            item_map: Dict[str, int] = {}
            interactions: Dict[Tuple[int, int], float] = defaultdict(float)

            for item_json in items:
                try:
                    data = json.loads(item_json)
                    uid = data["user_id"]
                    fid = data["function_id"]
                    weight = data.get("weight", 1.0)

                    if uid not in user_map:
                        user_map[uid] = len(user_map)
                    if fid not in item_map:
                        item_map[fid] = len(item_map)

                    key = (user_map[uid], item_map[fid])
                    interactions[key] += weight
                except Exception:
                    continue

            if not interactions:
                return csr_matrix((0, 0)), {}, {}

            rows, cols, vals = [], [], []
            for (u, i), v in interactions.items():
                rows.append(u)
                cols.append(i)
                vals.append(min(v, 10.0))  # Cap at 10

            matrix = csr_matrix(
                (vals, (rows, cols)),
                shape=(len(user_map), len(item_map)),
            )
            return matrix, user_map, item_map

        except Exception as e:
            logger.error(f"Failed to build interaction matrix: {e}")
            return csr_matrix((0, 0)), {}, {}

    async def train(self) -> bool:
        """Train ALS collaborative filtering model.

        Uses alternating least squares to factorize the interaction matrix
        into user and item latent factor matrices.
        """
        matrix, user_map, item_map = await self._build_interaction_matrix()

        if matrix.shape[0] < 2 or matrix.shape[1] < 2:
            logger.info("Not enough data to train recommendation model")
            return False

        n_users, n_items = matrix.shape
        k = min(self._latent_dims, min(n_users, n_items) - 1)

        if k < 1:
            return False

        rng = np.random.RandomState(42)
        user_factors = rng.normal(0, 0.1, (n_users, k))
        item_factors = rng.normal(0, 0.1, (n_items, k))

        # ALS iterations
        lambda_reg = 0.1
        n_iterations = 10

        for _ in range(n_iterations):
            # Update user factors
            for u in range(n_users):
                start = matrix.indptr[u]
                end = matrix.indptr[u + 1]
                if start == end:
                    continue
                items_idx = matrix.indices[start:end]
                ratings = matrix.data[start:end]

                A = item_factors[items_idx].T @ item_factors[items_idx] + lambda_reg * np.eye(k)
                b = item_factors[items_idx].T @ ratings
                user_factors[u] = np.linalg.solve(A, b)

            # Update item factors
            matrix_T = matrix.T.tocsr()
            for i in range(n_items):
                start = matrix_T.indptr[i]
                end = matrix_T.indptr[i + 1]
                if start == end:
                    continue
                users_idx = matrix_T.indices[start:end]
                ratings = matrix_T.data[start:end]

                A = user_factors[users_idx].T @ user_factors[users_idx] + lambda_reg * np.eye(k)
                b = user_factors[users_idx].T @ ratings
                item_factors[i] = np.linalg.solve(A, b)

        self._user_factors = user_factors
        self._item_factors = item_factors
        self._user_map = user_map
        self._item_map = item_map
        self._reverse_item_map = {v: k for k, v in item_map.items()}

        # Cache in Redis
        r = await self.get_redis()
        if r:
            try:
                import io
                buf = io.BytesIO()
                np.savez_compressed(buf, user=user_factors, item=item_factors)
                buf.seek(0)
                await r.set(self.FACTORS_USER_KEY, buf.read().hex(), ex=self.EXPIRY_DAYS * 86400)
                await r.set(
                    self.USER_MAP_KEY, json.dumps(user_map), ex=self.EXPIRY_DAYS * 86400
                )
                await r.set(
                    self.ITEM_MAP_KEY, json.dumps(item_map), ex=self.EXPIRY_DAYS * 86400
                )
            except Exception as e:
                logger.warning(f"Failed to cache factors: {e}")

        logger.info(
            f"Trained recommendation model: {n_users} users, {n_items} items, {k} dims"
        )
        return True

    async def recommend(
        self,
        user_id: str,
        limit: int = 20,
        exclude_ids: Optional[List[str]] = None,
    ) -> RecommendationResponse:
        """Get personalized recommendations for a user.

        Falls back to popularity-based for cold-start users.
        """
        exclude_set = set(exclude_ids or [])

        # Try collaborative filtering
        if (
            self._user_factors is not None
            and self._item_factors is not None
            and user_id in self._user_map
        ):
            user_idx = self._user_map[user_id]
            scores = self._user_factors[user_idx] @ self._item_factors.T

            # Build recommendations
            scored_items = []
            for item_idx, score in enumerate(scores):
                fid = self._reverse_item_map.get(item_idx)
                if fid and fid not in exclude_set:
                    # Normalize score to 0-1
                    norm_score = 1.0 / (1.0 + np.exp(-score))
                    scored_items.append((fid, float(norm_score)))

            scored_items.sort(key=lambda x: x[1], reverse=True)

            return RecommendationResponse(
                user_id=user_id,
                recommendations=[
                    RecommendationResult(
                        function_id=fid,
                        score=round(s, 3),
                        reason="Based on users with similar interests",
                        strategy="collaborative_filtering",
                    )
                    for fid, s in scored_items[:limit]
                ],
                strategy="collaborative_filtering",
            )

        # Fallback: popularity-based
        return await self._popularity_recommendations(user_id, limit, exclude_set)

    async def _popularity_recommendations(
        self, user_id: str, limit: int, exclude_set: set
    ) -> RecommendationResponse:
        """Popularity-based fallback for cold-start users."""
        r = await self.get_redis()
        if not r:
            return RecommendationResponse(
                user_id=user_id,
                recommendations=[],
                strategy="popularity",
            )

        try:
            popular = await r.zrevrange(self.POPULARITY_KEY, 0, limit + len(exclude_set) - 1, withscores=True)
            results = []
            for fid, score in popular:
                if fid not in exclude_set and len(results) < limit:
                    results.append(RecommendationResult(
                        function_id=fid,
                        score=min(1.0, score / 100.0),
                        reason="Popular among all users",
                        strategy="popularity",
                    ))
            return RecommendationResponse(
                user_id=user_id,
                recommendations=results,
                strategy="popularity",
            )
        except Exception as e:
            logger.error(f"Popularity recommendations failed: {e}")
            return RecommendationResponse(user_id=user_id, recommendations=[], strategy="error")

    async def close(self):
        if self._redis:
            await self._redis.close()
            self._redis = None


_engine: Optional[RecommendationEngine] = None


def get_recommendation_engine() -> RecommendationEngine:
    global _engine
    if _engine is None:
        _engine = RecommendationEngine()
    return _engine
