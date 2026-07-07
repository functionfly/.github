"""Tenant-isolated Redis operations for FlyMind ML services.

Ensures all Redis keys are namespaced by tenant to prevent cross-tenant data access.
"""

import json
import logging
from typing import Any, Dict, List, Optional, Tuple

import redis.asyncio as redis

from ...config import settings

logger = logging.getLogger(__name__)


class TenantRedisKeys:
    """Generates tenant-isolated Redis key names.
    
    Format: ml:{tenant_id}:{service}:{resource}:{id}
    
    This ensures complete tenant isolation in Redis operations.
    """

    COST_ANOMALY_STATS = "cost_anomaly:stats"
    COST_ANOMALY_RECORDS = "cost_anomaly:records"
    COST_ANOMALY_MEMORY = "cost_anomaly:memory"
    THOMPSON_ARMS = "thompson:arms"
    THOMPSON_OUTCOMES = "thompson:outcomes"
    PREWARM_HISTORY = "prewarming:history"
    PREWARM_MODEL = "prewarming:model"
    REC_INTERACTIONS = "recommendations:interactions"
    REC_FACTORS_USER = "recommendations:users"
    REC_FACTORS_ITEM = "recommendations:items"
    REC_POPULARITY = "recommendations:popularity"
    REC_USER_MAP = "recommendations:user_map"
    REC_ITEM_MAP = "recommendations:item_map"

    @classmethod
    def cost_anomaly_stats(cls, tenant_id: str, function_id: str) -> str:
        return f"ml:{tenant_id}:{cls.COST_ANOMALY_STATS}:{function_id}"

    @classmethod
    def cost_anomaly_records(cls, tenant_id: str) -> str:
        return f"ml:{tenant_id}:{cls.COST_ANOMALY_RECORDS}"

    @classmethod
    def cost_anomaly_memory(cls, tenant_id: str, function_id: str) -> str:
        return f"ml:{tenant_id}:{cls.COST_ANOMALY_MEMORY}:{function_id}"

    @classmethod
    def thompson_arms(cls, tenant_id: str, function_id: str) -> str:
        return f"ml:{tenant_id}:{cls.THOMPSON_ARMS}:{function_id}"

    @classmethod
    def thompson_outcomes(cls, tenant_id: str, function_id: str) -> str:
        return f"ml:{tenant_id}:{cls.THOMPSON_OUTCOMES}:{function_id}"

    @classmethod
    def prewarm_history(cls, tenant_id: str, function_id: str) -> str:
        return f"ml:{tenant_id}:{cls.PREWARM_HISTORY}:{function_id}"

    @classmethod
    def prewarm_model(cls, tenant_id: str, function_id: str) -> str:
        return f"ml:{tenant_id}:{cls.PREWARM_MODEL}:{function_id}"

    @classmethod
    def rec_interactions(cls, tenant_id: str) -> str:
        return f"ml:{tenant_id}:{cls.REC_INTERACTIONS}"

    @classmethod
    def rec_factors_user(cls, tenant_id: str) -> str:
        return f"ml:{tenant_id}:{cls.REC_FACTORS_USER}"

    @classmethod
    def rec_factors_item(cls, tenant_id: str) -> str:
        return f"ml:{tenant_id}:{cls.REC_FACTORS_ITEM}"

    @classmethod
    def rec_popularity(cls, tenant_id: str) -> str:
        return f"ml:{tenant_id}:{cls.REC_POPULARITY}"

    @classmethod
    def rec_user_map(cls, tenant_id: str) -> str:
        return f"ml:{tenant_id}:{cls.REC_USER_MAP}"

    @classmethod
    def rec_item_map(cls, tenant_id: str) -> str:
        return f"ml:{tenant_id}:{cls.REC_ITEM_MAP}"

    @classmethod
    def validate_tenant_access(cls, api_key_tenant_id: str, requested_tenant_id: str) -> bool:
        """Validate that the API key's tenant matches the requested tenant.
        
        Args:
            api_key_tenant_id: The tenant_id from the API key
            requested_tenant_id: The tenant_id in the request path
            
        Returns:
            True if access is allowed
        """
        return api_key_tenant_id == requested_tenant_id


class TenantScopedRedisClient:
    """Redis client wrapper that enforces tenant isolation.
    
    All operations automatically include tenant prefixing and validation.
    """

    def __init__(self, tenant_id: str):
        """Initialize with tenant ID.
        
        Args:
            tenant_id: The tenant ID for all operations
        """
        self._tenant_id = tenant_id
        self._redis: Optional[redis.Redis] = None

    @property
    def tenant_id(self) -> str:
        return self._tenant_id

    async def get_redis(self) -> Optional[redis.Redis]:
        """Get Redis connection."""
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

    async def close(self):
        """Close Redis connection."""
        if self._redis:
            await self._redis.close()
            self._redis = None

    async def get_cost_stats(self, function_id: str) -> Optional[Dict[str, Any]]:
        """Get cost anomaly stats for a function."""
        r = await self.get_redis()
        if not r:
            return None
        key = TenantRedisKeys.cost_anomaly_stats(self._tenant_id, function_id)
        try:
            data = await r.get(key)
            if data:
                return json.loads(data)
        except Exception as e:
            logger.warning(f"Failed to get cost stats for {function_id}: {e}")
        return None

    async def set_cost_stats(self, function_id: str, stats: Dict[str, Any], ttl_seconds: int) -> bool:
        """Set cost anomaly stats for a function."""
        r = await self.get_redis()
        if not r:
            return False
        key = TenantRedisKeys.cost_anomaly_stats(self._tenant_id, function_id)
        try:
            await r.set(key, json.dumps(stats), ex=ttl_seconds)
            return True
        except Exception as e:
            logger.error(f"Failed to set cost stats for {function_id}: {e}")
            return False

    async def get_cost_anomalies(self, limit: int = 50) -> List[Dict[str, Any]]:
        """Get cost anomaly records for tenant."""
        r = await self.get_redis()
        if not r:
            return []
        key = TenantRedisKeys.cost_anomaly_records(self._tenant_id)
        try:
            items = await r.zrevrange(key, 0, limit * 2)
            results = []
            for item in items:
                try:
                    results.append(json.loads(item))
                except Exception:
                    continue
            return results[:limit]
        except Exception as e:
            logger.error(f"Failed to get cost anomalies: {e}")
            return []

    async def add_cost_anomaly(self, anomaly_json: str, score: float, ttl_seconds: int) -> bool:
        """Add a cost anomaly record."""
        r = await self.get_redis()
        if not r:
            return False
        key = TenantRedisKeys.cost_anomaly_records(self._tenant_id)
        try:
            await r.zadd(key, {anomaly_json: score})
            await r.expire(key, ttl_seconds)
            return True
        except Exception as e:
            logger.error(f"Failed to add cost anomaly: {e}")
            return False

    async def get_thompson_arms(self, function_id: str) -> Optional[Dict[str, Dict[str, Any]]]:
        """Get Thompson Sampling arms for a function."""
        r = await self.get_redis()
        if not r:
            return None
        key = TenantRedisKeys.thompson_arms(self._tenant_id, function_id)
        try:
            data = await r.get(key)
            if data:
                return json.loads(data)
        except Exception as e:
            logger.warning(f"Failed to get Thompson arms for {function_id}: {e}")
        return None

    async def set_thompson_arms(
        self, function_id: str, arms: Dict[str, Dict[str, Any]], ttl_seconds: int
    ) -> bool:
        """Set Thompson Sampling arms for a function."""
        r = await self.get_redis()
        if not r:
            return False
        key = TenantRedisKeys.thompson_arms(self._tenant_id, function_id)
        try:
            await r.set(key, json.dumps(arms), ex=ttl_seconds)
            return True
        except Exception as e:
            logger.error(f"Failed to set Thompson arms for {function_id}: {e}")
            return False

    async def add_thompson_outcome(self, function_id: str, outcome_json: str) -> bool:
        """Add a Thompson Sampling outcome."""
        r = await self.get_redis()
        if not r:
            return False
        key = TenantRedisKeys.thompson_outcomes(self._tenant_id, function_id)
        try:
            await r.lpush(key, outcome_json)
            await r.ltrim(key, 0, 999)
            await r.expire(key, 86400 * 7)
            return True
        except Exception as e:
            logger.error(f"Failed to add Thompson outcome: {e}")
            return False

    async def get_prewarm_history(self, function_id: str, hours: int = 168) -> List[Tuple[str, float]]:
        """Get prewarming history for a function.
        
        Returns:
            List of (data_json, score) tuples sorted by timestamp
        """
        r = await self.get_redis()
        if not r:
            return []
        key = TenantRedisKeys.prewarm_history(self._tenant_id, function_id)
        try:
            min_score = (datetime.utcnow() - timedelta(hours=hours)).timestamp()
            items = await r.zrangebyscore(key, min_score, "+inf", withscores=True)
            return items
        except Exception as e:
            logger.error(f"Failed to get prewarm history: {e}")
            return []

    async def add_prewarm_request(self, function_id: str, data_json: str, ttl_hours: int = 168) -> bool:
        """Add a prewarming request record."""
        r = await self.get_redis()
        if not r:
            return False
        key = TenantRedisKeys.prewarm_history(self._tenant_id, function_id)
        try:
            score = datetime.utcnow().timestamp()
            await r.zadd(key, {data_json: score})
            await r.expire(key, ttl_hours * 3600)
            return True
        except Exception as e:
            logger.error(f"Failed to add prewarm request: {e}")
            return False

    async def get_rec_interactions(self, limit: int = -1) -> List[str]:
        """Get recommendation interactions for tenant."""
        r = await self.get_redis()
        if not r:
            return []
        key = TenantRedisKeys.rec_interactions(self._tenant_id)
        try:
            if limit < 0:
                return await r.zrange(key, 0, -1)
            return await r.zrange(key, 0, limit)
        except Exception as e:
            logger.error(f"Failed to get interactions: {e}")
            return []

    async def add_rec_interaction(self, entry_json: str, score: float, ttl_seconds: int) -> bool:
        """Add a recommendation interaction."""
        r = await self.get_redis()
        if not r:
            return False
        key = TenantRedisKeys.rec_interactions(self._tenant_id)
        try:
            await r.zadd(key, {entry_json: score})
            await r.expire(key, ttl_seconds)
            return True
        except Exception as e:
            logger.error(f"Failed to add interaction: {e}")
            return False

    async def incr_rec_popularity(self, function_id: str, weight: float) -> bool:
        """Increment recommendation popularity counter."""
        r = await self.get_redis()
        if not r:
            return False
        key = TenantRedisKeys.rec_popularity(self._tenant_id)
        try:
            await r.zincrby(key, weight, function_id)
            return True
        except Exception as e:
            logger.error(f"Failed to incr popularity: {e}")
            return False

    async def get_rec_popularity(self, function_id: str, limit: int = 20) -> List[Tuple[str, float]]:
        """Get popular functions for tenant."""
        r = await self.get_redis()
        if not r:
            return []
        key = TenantRedisKeys.rec_popularity(self._tenant_id)
        try:
            return await r.zrevrange(key, 0, limit + 9, withscores=True)
        except Exception as e:
            logger.error(f"Failed to get popularity: {e}")
            return []

    async def set_rec_factors(self, user_factors_hex: str, item_factors_hex: str, 
                               user_map_json: str, item_map_json: str, 
                               ttl_seconds: int) -> bool:
        """Cache recommendation factors."""
        r = await self.get_redis()
        if not r:
            return False
        try:
            await r.set(TenantRedisKeys.rec_factors_user(self._tenant_id), 
                       user_factors_hex, ex=ttl_seconds)
            await r.set(TenantRedisKeys.rec_factors_item(self._tenant_id), 
                       item_factors_hex, ex=ttl_seconds)
            await r.set(TenantRedisKeys.rec_user_map(self._tenant_id), 
                       user_map_json, ex=ttl_seconds)
            await r.set(TenantRedisKeys.rec_item_map(self._tenant_id), 
                       item_map_json, ex=ttl_seconds)
            return True
        except Exception as e:
            logger.warning(f"Failed to cache factors: {e}")
            return False

    async def get_rec_factors(self) -> Optional[Dict[str, Any]]:
        """Get cached recommendation factors."""
        r = await self.get_redis()
        if not r:
            return None
        try:
            user_factors = await r.get(TenantRedisKeys.rec_factors_user(self._tenant_id))
            item_factors = await r.get(TenantRedisKeys.rec_factors_item(self._tenant_id))
            user_map = await r.get(TenantRedisKeys.rec_user_map(self._tenant_id))
            item_map = await r.get(TenantRedisKeys.rec_item_map(self._tenant_id))
            
            if all([user_factors, item_factors, user_map, item_map]):
                return {
                    "user_factors": user_factors,
                    "item_factors": item_factors,
                    "user_map": json.loads(user_map),
                    "item_map": json.loads(item_map),
                }
        except Exception as e:
            logger.warning(f"Failed to get factors: {e}")
        return None

    async def invalidate_rec_factors(self) -> bool:
        """Invalidate cached recommendation factors."""
        r = await self.get_redis()
        if not r:
            return False
        try:
            await r.delete(
                TenantRedisKeys.rec_factors_user(self._tenant_id),
                TenantRedisKeys.rec_factors_item(self._tenant_id),
            )
            return True
        except Exception as e:
            logger.error(f"Failed to invalidate factors: {e}")
            return False


from datetime import datetime, timedelta
