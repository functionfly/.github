"""Database persistence for Economic Memory.

Stores cost-quality metrics in PostgreSQL for long-term analysis and
tenant-level insights.
"""

import logging
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Any
from dataclasses import asdict

import asyncpg

from ...config import settings
from . import CostQualityScore, ExecutionRecord, ProviderType

logger = logging.getLogger(__name__)


class EconomicMemoryRepository:
    """Repository for persisting economic memory data to PostgreSQL."""
    
    def __init__(self, pool: Optional[asyncpg.Pool] = None):
        self._pool = pool
        self._dsn = settings.database_url
    
    async def _get_pool(self) -> asyncpg.Pool:
        """Get or create connection pool."""
        if self._pool is None:
            self._pool = await asyncpg.create_pool(self._dsn, min_size=1, max_size=10)
        return self._pool
    
    async def save_execution_record(self, record: ExecutionRecord) -> bool:
        """Save an execution record to the database."""
        try:
            pool = await self._get_pool()
            async with pool.acquire() as conn:
                await conn.execute(
                    """
                    INSERT INTO economic_memory_executions (
                        execution_id, provider, model, tenant_id, function_id,
                        input_tokens, output_tokens, total_tokens, cost_usd,
                        latency_ms, success, error_type, output_quality_score,
                        user_rating, timestamp
                    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
                    ON CONFLICT (execution_id) DO NOTHING
                    """,
                    record.execution_id,
                    record.provider.value,
                    record.model,
                    record.tenant_id,
                    record.function_id,
                    record.input_tokens,
                    record.output_tokens,
                    record.total_tokens,
                    record.cost_usd,
                    record.latency_ms,
                    record.success,
                    record.error_type,
                    record.output_quality_score,
                    record.user_rating,
                    record.timestamp,
                )
            return True
        except Exception as e:
            logger.error(f"Failed to save execution record: {e}")
            return False
    
    async def save_cost_quality_score(self, score: CostQualityScore) -> bool:
        """Save or update a cost-quality score to the database."""
        try:
            pool = await self._get_pool()
            async with pool.acquire() as conn:
                await conn.execute(
                    """
                    INSERT INTO economic_memory_scores (
                        provider, model, avg_cost_per_1k_tokens, avg_cost_per_request,
                        quality_score, response_time_score, token_efficiency_score,
                        success_rate, cost_quality_index, total_executions,
                        total_cost_usd, total_tokens, last_updated
                    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
                    ON CONFLICT (provider, model) DO UPDATE SET
                        avg_cost_per_1k_tokens = EXCLUDED.avg_cost_per_1k_tokens,
                        avg_cost_per_request = EXCLUDED.avg_cost_per_request,
                        quality_score = EXCLUDED.quality_score,
                        response_time_score = EXCLUDED.response_time_score,
                        token_efficiency_score = EXCLUDED.token_efficiency_score,
                        success_rate = EXCLUDED.success_rate,
                        cost_quality_index = EXCLUDED.cost_quality_index,
                        total_executions = EXCLUDED.total_executions,
                        total_cost_usd = EXCLUDED.total_cost_usd,
                        total_tokens = EXCLUDED.total_tokens,
                        last_updated = EXCLUDED.last_updated
                    """,
                    score.provider.value,
                    score.model,
                    score.avg_cost_per_1k_tokens,
                    score.avg_cost_per_request,
                    score.quality_score,
                    score.response_time_score,
                    score.token_efficiency_score,
                    score.success_rate,
                    score.cost_quality_index,
                    score.total_executions,
                    score.total_cost_usd,
                    score.total_tokens,
                    score.last_updated,
                )
            return True
        except Exception as e:
            logger.error(f"Failed to save cost-quality score: {e}")
            return False
    
    async def load_cost_quality_scores(self) -> List[CostQualityScore]:
        """Load all cost-quality scores from the database."""
        try:
            pool = await self._get_pool()
            async with pool.acquire() as conn:
                rows = await conn.fetch(
                    """
                    SELECT provider, model, avg_cost_per_1k_tokens, avg_cost_per_request,
                           quality_score, response_time_score, token_efficiency_score,
                           success_rate, cost_quality_index, total_executions,
                           total_cost_usd, total_tokens, last_updated
                    FROM economic_memory_scores
                    """
                )
                
                scores = []
                for row in rows:
                    try:
                        score = CostQualityScore(
                            provider=ProviderType(row["provider"]),
                            model=row["model"],
                            avg_cost_per_1k_tokens=row["avg_cost_per_1k_tokens"],
                            avg_cost_per_request=row["avg_cost_per_request"],
                            quality_score=row["quality_score"],
                            response_time_score=row["response_time_score"],
                            token_efficiency_score=row["token_efficiency_score"],
                            success_rate=row["success_rate"],
                            cost_quality_index=row["cost_quality_index"],
                            total_executions=row["total_executions"],
                            total_cost_usd=row["total_cost_usd"],
                            total_tokens=row["total_tokens"],
                            last_updated=row["last_updated"],
                        )
                        scores.append(score)
                    except Exception as e:
                        logger.warning(f"Failed to parse score row: {e}")
                
                return scores
        except Exception as e:
            logger.error(f"Failed to load cost-quality scores: {e}")
            return []
    
    async def get_tenant_cost_summary(
        self,
        tenant_id: str,
        days: int = 7,
    ) -> Dict[str, Any]:
        """Get cost summary for a tenant over the specified period."""
        try:
            pool = await self._get_pool()
            since = datetime.utcnow() - timedelta(days=days)
            
            async with pool.acquire() as conn:
                row = await conn.fetchrow(
                    """
                    SELECT 
                        COUNT(*) as executions,
                        SUM(cost_usd) as total_cost,
                        SUM(total_tokens) as total_tokens,
                        AVG(cost_usd) as avg_cost,
                        AVG(latency_ms) as avg_latency,
                        SUM(CASE WHEN success THEN 1 ELSE 0 END)::float / COUNT(*) as success_rate
                    FROM economic_memory_executions
                    WHERE tenant_id = $1 AND timestamp >= $2
                    """,
                    tenant_id,
                    since,
                )
                
                if row:
                    return {
                        "tenant_id": tenant_id,
                        "period_days": days,
                        "executions": row["executions"] or 0,
                        "total_cost_usd": row["total_cost"] or 0.0,
                        "total_tokens": row["total_tokens"] or 0,
                        "avg_cost_per_execution": row["avg_cost"] or 0.0,
                        "avg_latency_ms": row["avg_latency"] or 0.0,
                        "success_rate": row["success_rate"] or 0.0,
                    }
                
                return {
                    "tenant_id": tenant_id,
                    "period_days": days,
                    "executions": 0,
                    "total_cost_usd": 0.0,
                    "total_tokens": 0,
                    "avg_cost_per_execution": 0.0,
                    "avg_latency_ms": 0.0,
                    "success_rate": 0.0,
                }
        except Exception as e:
            logger.error(f"Failed to get tenant cost summary: {e}")
            return {}
    
    async def get_provider_comparison(
        self,
        days: int = 7,
    ) -> List[Dict[str, Any]]:
        """Get provider comparison statistics."""
        try:
            pool = await self._get_pool()
            since = datetime.utcnow() - timedelta(days=days)
            
            async with pool.acquire() as conn:
                rows = await conn.fetch(
                    """
                    SELECT 
                        provider,
                        model,
                        COUNT(*) as executions,
                        SUM(cost_usd) as total_cost,
                        AVG(cost_usd) as avg_cost,
                        SUM(total_tokens) as total_tokens,
                        AVG(latency_ms) as avg_latency,
                        SUM(CASE WHEN success THEN 1 ELSE 0 END)::float / COUNT(*) as success_rate,
                        AVG(output_quality_score) as avg_quality
                    FROM economic_memory_executions
                    WHERE timestamp >= $1
                    GROUP BY provider, model
                    ORDER BY total_cost DESC
                    """,
                    since,
                )
                
                results = []
                for row in rows:
                    results.append({
                        "provider": row["provider"],
                        "model": row["model"],
                        "executions": row["executions"],
                        "total_cost_usd": row["total_cost"],
                        "avg_cost_per_execution": row["avg_cost"],
                        "total_tokens": row["total_tokens"],
                        "avg_latency_ms": row["avg_latency"],
                        "success_rate": row["success_rate"],
                        "avg_quality_score": row["avg_quality"] or 0.0,
                    })
                
                return results
        except Exception as e:
            logger.error(f"Failed to get provider comparison: {e}")
            return []
    
    async def get_daily_cost_trend(
        self,
        tenant_id: Optional[str] = None,
        days: int = 30,
    ) -> List[Dict[str, Any]]:
        """Get daily cost trend for charting."""
        try:
            pool = await self._get_pool()
            since = datetime.utcnow() - timedelta(days=days)
            
            query = """
                SELECT 
                    DATE(timestamp) as date,
                    COUNT(*) as executions,
                    SUM(cost_usd) as daily_cost,
                    SUM(total_tokens) as daily_tokens,
                    AVG(latency_ms) as avg_latency,
                    SUM(CASE WHEN success THEN 1 ELSE 0 END)::float / COUNT(*) as success_rate
                FROM economic_memory_executions
                WHERE timestamp >= $1
            """
            params = [since]
            
            if tenant_id:
                query += " AND tenant_id = $2"
                params.append(tenant_id)
            
            query += " GROUP BY DATE(timestamp) ORDER BY date ASC"
            
            async with pool.acquire() as conn:
                rows = await conn.fetch(query, *params)
                
                results = []
                for row in rows:
                    results.append({
                        "date": row["date"].isoformat() if row["date"] else None,
                        "executions": row["executions"],
                        "cost_usd": row["daily_cost"],
                        "tokens": row["daily_tokens"],
                        "avg_latency_ms": row["avg_latency"],
                        "success_rate": row["success_rate"],
                    })
                
                return results
        except Exception as e:
            logger.error(f"Failed to get daily cost trend: {e}")
            return []


# Migration SQL (to be applied to database)
MIGRATION_SQL = """
-- Economic Memory tables for Phase 3

-- Execution records table
CREATE TABLE IF NOT EXISTS economic_memory_executions (
    id SERIAL PRIMARY KEY,
    execution_id UUID UNIQUE NOT NULL,
    provider VARCHAR(50) NOT NULL,
    model VARCHAR(100) NOT NULL,
    tenant_id VARCHAR(100),
    function_id VARCHAR(100),
    
    -- Cost metrics
    input_tokens INTEGER DEFAULT 0,
    output_tokens INTEGER DEFAULT 0,
    total_tokens INTEGER DEFAULT 0,
    cost_usd DECIMAL(12, 8) DEFAULT 0,
    
    -- Quality metrics
    latency_ms DECIMAL(10, 2) DEFAULT 0,
    success BOOLEAN DEFAULT TRUE,
    error_type VARCHAR(100),
    output_quality_score DECIMAL(3, 2),
    user_rating DECIMAL(2, 1),
    
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_economic_exec_timestamp 
    ON economic_memory_executions(timestamp);
CREATE INDEX IF NOT EXISTS idx_economic_exec_tenant 
    ON economic_memory_executions(tenant_id, timestamp);
CREATE INDEX IF NOT EXISTS idx_economic_exec_provider 
    ON economic_memory_executions(provider, model, timestamp);

-- Aggregated scores table
CREATE TABLE IF NOT EXISTS economic_memory_scores (
    id SERIAL PRIMARY KEY,
    provider VARCHAR(50) NOT NULL,
    model VARCHAR(100) NOT NULL,
    
    -- Cost metrics
    avg_cost_per_1k_tokens DECIMAL(12, 8) DEFAULT 0,
    avg_cost_per_request DECIMAL(12, 8) DEFAULT 0,
    
    -- Quality metrics (0-1 scale)
    quality_score DECIMAL(3, 2) DEFAULT 0,
    response_time_score DECIMAL(3, 2) DEFAULT 0,
    token_efficiency_score DECIMAL(5, 2) DEFAULT 0,
    success_rate DECIMAL(3, 2) DEFAULT 1.0,
    
    -- Composite score
    cost_quality_index DECIMAL(5, 2) DEFAULT 0,
    
    -- Totals
    total_executions INTEGER DEFAULT 0,
    total_cost_usd DECIMAL(12, 4) DEFAULT 0,
    total_tokens BIGINT DEFAULT 0,
    
    last_updated TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(provider, model)
);

-- Index for score lookup
CREATE INDEX IF NOT EXISTS idx_economic_scores_cqi 
    ON economic_memory_scores(cost_quality_index DESC);
"""


async def run_migration():
    """Run the database migration."""
    try:
        conn = await asyncpg.connect(settings.database_url)
        try:
            await conn.execute(MIGRATION_SQL)
            logger.info("Economic memory migration completed successfully")
        finally:
            await conn.close()
    except Exception as e:
        logger.error(f"Migration failed: {e}")
        raise


# Global repository instance
_repository: Optional[EconomicMemoryRepository] = None


def get_repository() -> EconomicMemoryRepository:
    """Get the global repository instance."""
    global _repository
    if _repository is None:
        _repository = EconomicMemoryRepository()
    return _repository
