"""Economic Memory service for cost-per-quality tracking.

Tracks cost vs output quality per execution to enable cost-intelligent
model selection and routing decisions. This is Phase 3 of the FlyMind
AI Service implementation.
"""

import logging
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Tuple
from dataclasses import dataclass, field
from enum import Enum
import asyncio
from collections import defaultdict

from ...config import settings
from ...models.schemas import ProviderType

logger = logging.getLogger(__name__)


class QualityMetric(str, Enum):
    """Types of quality metrics tracked."""
    RESPONSE_TIME = "response_time"
    TOKEN_EFFICIENCY = "token_efficiency"
    OUTPUT_QUALITY = "output_quality"
    SUCCESS_RATE = "success_rate"
    USER_RATING = "user_rating"


@dataclass
class CostQualityScore:
    """Score combining cost and quality metrics for a model/provider."""
    provider: ProviderType
    model: str
    
    # Cost metrics
    avg_cost_per_1k_tokens: float = 0.0
    avg_cost_per_request: float = 0.0
    
    # Quality metrics (0-1 scale)
    quality_score: float = 0.0
    response_time_score: float = 0.0
    token_efficiency_score: float = 0.0
    success_rate: float = 1.0
    
    # Composite score (higher = better value)
    cost_quality_index: float = 0.0
    
    # Metadata
    total_executions: int = 0
    total_cost_usd: float = 0.0
    total_tokens: int = 0
    last_updated: datetime = field(default_factory=datetime.utcnow)
    
    def calculate_cqi(self) -> float:
        """Calculate the Cost-Quality Index (CQI).
        
        CQI = (quality_score * response_time_score * success_rate) / avg_cost_per_1k_tokens
        Higher is better - more quality per dollar spent.
        """
        if self.avg_cost_per_1k_tokens <= 0:
            return 0.0
        
        # Base quality composite
        quality_composite = (
            self.quality_score * 0.4 +
            self.response_time_score * 0.3 +
            self.token_efficiency_score * 0.2 +
            self.success_rate * 0.1
        )
        
        # Cost-quality index: quality per dollar, normalized
        # Scale to 0-100 range for readability
        raw_cqi = (quality_composite * 100) / (self.avg_cost_per_1k_tokens * 1000)
        self.cost_quality_index = min(100.0, max(0.0, raw_cqi))
        return self.cost_quality_index


@dataclass
class ExecutionRecord:
    """Record of a single execution for cost-quality tracking."""
    execution_id: str
    provider: ProviderType
    model: str
    
    # Cost data
    input_tokens: int = 0
    output_tokens: int = 0
    total_tokens: int = 0
    cost_usd: float = 0.0
    
    # Quality data
    latency_ms: float = 0.0
    success: bool = True
    error_type: Optional[str] = None
    output_quality_score: Optional[float] = None  # 0-1, from validation/analysis
    user_rating: Optional[float] = None  # 0-5, from user feedback
    
    # Context
    tenant_id: Optional[str] = None
    function_id: Optional[str] = None
    timestamp: datetime = field(default_factory=datetime.utcnow)
    
    # Calculated fields
    cost_per_1k_tokens: float = field(init=False)
    
    def __post_init__(self):
        if self.total_tokens > 0:
            self.cost_per_1k_tokens = (self.cost_usd / self.total_tokens) * 1000
        else:
            self.cost_per_1k_tokens = 0.0


class EconomicMemory:
    """In-memory store for cost-quality metrics with persistence."""
    
    def __init__(self):
        self._scores: Dict[Tuple[str, str], CostQualityScore] = {}
        self._recent_executions: List[ExecutionRecord] = []
        self._lock = asyncio.Lock()
        self._max_recent = 1000  # Keep last 1000 executions in memory
        
    async def record_execution(self, record: ExecutionRecord) -> None:
        """Record an execution and update cost-quality scores."""
        async with self._lock:
            # Add to recent executions
            self._recent_executions.append(record)
            if len(self._recent_executions) > self._max_recent:
                self._recent_executions.pop(0)
            
            # Update or create score for this provider/model
            key = (record.provider.value, record.model)
            
            if key not in self._scores:
                self._scores[key] = CostQualityScore(
                    provider=record.provider,
                    model=record.model,
                )
            
            score = self._scores[key]
            
            # Update running averages
            n = score.total_executions
            score.total_executions += 1
            score.total_cost_usd += record.cost_usd
            score.total_tokens += record.total_tokens
            
            # Update cost averages
            if record.total_tokens > 0:
                new_cost_per_1k = (record.cost_usd / record.total_tokens) * 1000
                score.avg_cost_per_1k_tokens = (
                    (score.avg_cost_per_1k_tokens * n + new_cost_per_1k) / (n + 1)
                )
            
            score.avg_cost_per_request = (
                (score.avg_cost_per_request * n + record.cost_usd) / (n + 1)
            )
            
            # Update quality metrics
            # Latency score: faster is better (target < 1000ms)
            latency_score = max(0.0, 1.0 - (record.latency_ms / 1000.0))
            score.response_time_score = (
                (score.response_time_score * n + latency_score) / (n + 1)
            )
            
            # Success rate
            success_val = 1.0 if record.success else 0.0
            score.success_rate = (
                (score.success_rate * n + success_val) / (n + 1)
            )
            
            # Token efficiency (output/input ratio, higher is better)
            if record.input_tokens > 0:
                efficiency = record.output_tokens / record.input_tokens
                if n == 0:
                    score.token_efficiency_score = efficiency
                else:
                    score.token_efficiency_score = (
                        (score.token_efficiency_score * n + efficiency) / (n + 1)
                    )
            
            # Output quality score if available
            if record.output_quality_score is not None:
                if n == 0:
                    score.quality_score = record.output_quality_score
                else:
                    score.quality_score = (
                        (score.quality_score * n + record.output_quality_score) / (n + 1)
                    )
            
            # Recalculate CQI
            score.calculate_cqi()
            score.last_updated = datetime.utcnow()
            
            logger.debug(
                f"Recorded execution for {record.provider.value}/{record.model}: "
                f"cost=${record.cost_usd:.6f}, tokens={record.total_tokens}, "
                f"CQI={score.cost_quality_index:.2f}"
            )
    
    async def get_score(
        self, 
        provider: ProviderType, 
        model: str
    ) -> Optional[CostQualityScore]:
        """Get the cost-quality score for a provider/model."""
        async with self._lock:
            key = (provider.value, model)
            return self._scores.get(key)
    
    async def get_all_scores(self) -> List[CostQualityScore]:
        """Get all cost-quality scores."""
        async with self._lock:
            return list(self._scores.values())
    
    async def get_best_value_provider(
        self,
        providers: Optional[List[ProviderType]] = None,
        min_executions: int = 10,
    ) -> Optional[CostQualityScore]:
        """Get the provider with the best cost-quality index.
        
        Args:
            providers: Optional list of providers to consider
            min_executions: Minimum executions required for consideration
            
        Returns:
            Best value provider or None if no data
        """
        async with self._lock:
            candidates = []
            
            for score in self._scores.values():
                if score.total_executions < min_executions:
                    continue
                if providers and score.provider not in providers:
                    continue
                candidates.append(score)
            
            if not candidates:
                return None
            
            # Sort by cost-quality index (descending)
            candidates.sort(key=lambda s: s.cost_quality_index, reverse=True)
            return candidates[0]
    
    async def get_recommendations(
        self,
        target_quality_threshold: float = 0.7,
        max_budget_usd: Optional[float] = None,
    ) -> List[Dict]:
        """Get model recommendations based on cost-quality analysis.
        
        Returns:
            List of recommendations sorted by value
        """
        async with self._lock:
            recommendations = []
            
            for score in self._scores.values():
                if score.total_executions < 5:  # Need some data
                    continue
                
                rec = {
                    "provider": score.provider.value,
                    "model": score.model,
                    "cost_quality_index": score.cost_quality_index,
                    "avg_cost_per_1k_tokens": score.avg_cost_per_1k_tokens,
                    "quality_score": score.quality_score,
                    "success_rate": score.success_rate,
                    "total_executions": score.total_executions,
                    "recommendation": "neutral",
                }
                
                # Determine recommendation
                if score.cost_quality_index >= 50 and score.quality_score >= target_quality_threshold:
                    rec["recommendation"] = "highly_recommended"
                elif score.cost_quality_index >= 30:
                    rec["recommendation"] = "recommended"
                elif score.cost_quality_index < 10:
                    rec["recommendation"] = "avoid"
                
                if max_budget_usd and score.avg_cost_per_request > max_budget_usd:
                    rec["recommendation"] = "over_budget"
                
                recommendations.append(rec)
            
            # Sort by CQI descending
            recommendations.sort(key=lambda r: r["cost_quality_index"], reverse=True)
            return recommendations
    
    async def get_recent_executions(
        self,
        provider: Optional[ProviderType] = None,
        model: Optional[str] = None,
        since: Optional[datetime] = None,
        limit: int = 100,
    ) -> List[ExecutionRecord]:
        """Get recent execution records with optional filtering."""
        async with self._lock:
            results = []
            
            for record in reversed(self._recent_executions):  # Most recent first
                if provider and record.provider != provider:
                    continue
                if model and record.model != model:
                    continue
                if since and record.timestamp < since:
                    continue
                
                results.append(record)
                if len(results) >= limit:
                    break
            
            return results
    
    async def get_cost_breakdown(self, days: int = 7) -> Dict:
        """Get cost breakdown by provider for the last N days."""
        async with self._lock:
            since = datetime.utcnow() - timedelta(days=days)
            
            breakdown = defaultdict(lambda: {
                "total_cost": 0.0,
                "total_tokens": 0,
                "executions": 0,
                "avg_quality": 0.0,
            })
            
            for record in self._recent_executions:
                if record.timestamp < since:
                    continue
                
                key = f"{record.provider.value}/{record.model}"
                entry = breakdown[key]
                entry["total_cost"] += record.cost_usd
                entry["total_tokens"] += record.total_tokens
                entry["executions"] += 1
                
                if record.output_quality_score:
                    # Running average of quality
                    n = entry["executions"]
                    entry["avg_quality"] = (
                        (entry["avg_quality"] * (n - 1) + record.output_quality_score) / n
                    )
            
            return {
                "period_days": days,
                "provider_breakdown": dict(breakdown),
                "total_cost": sum(b["total_cost"] for b in breakdown.values()),
                "total_executions": sum(b["executions"] for b in breakdown.values()),
            }
    
    async def suggest_model_switch(
        self,
        current_provider: ProviderType,
        current_model: str,
        target_quality: Optional[float] = None,
    ) -> Optional[Dict]:
        """Suggest a model switch if there's a better value option.
        
        Returns:
            Suggestion dict or None if current is optimal
        """
        async with self._lock:
            current_key = (current_provider.value, current_model)
            current_score = self._scores.get(current_key)
            
            if not current_score or current_score.total_executions < 5:
                return None
            
            # Find alternatives with similar or better quality
            best_alternative = None
            best_savings = 0.0
            
            for score in self._scores.values():
                if score.provider == current_provider and score.model == current_model:
                    continue
                if score.total_executions < 10:  # Need more data for alternatives
                    continue
                
                # Check quality meets target
                if target_quality and score.quality_score < target_quality:
                    continue
                
                # Calculate potential savings
                cost_diff = current_score.avg_cost_per_request - score.avg_cost_per_request
                quality_diff = score.quality_score - current_score.quality_score
                cqi_diff = score.cost_quality_index - current_score.cost_quality_index
                
                # Worth switching if CQI is better and quality is comparable
                if cqi_diff > 5 and quality_diff >= -0.1:
                    savings_pct = (cost_diff / current_score.avg_cost_per_request) * 100
                    if savings_pct > best_savings:
                        best_savings = savings_pct
                        best_alternative = {
                            "current_provider": current_provider.value,
                            "current_model": current_model,
                            "current_cost_per_request": current_score.avg_cost_per_request,
                            "current_cqi": current_score.cost_quality_index,
                            "suggested_provider": score.provider.value,
                            "suggested_model": score.model,
                            "suggested_cost_per_request": score.avg_cost_per_request,
                            "suggested_cqi": score.cost_quality_index,
                            "potential_savings_percent": round(savings_pct, 1),
                            "quality_delta": round(quality_diff, 2),
                            "confidence": "high" if score.total_executions > 50 else "medium",
                        }
            
            return best_alternative


# Global instance
_economic_memory: Optional[EconomicMemory] = None


def get_economic_memory() -> EconomicMemory:
    """Get the global economic memory instance."""
    global _economic_memory
    if _economic_memory is None:
        _economic_memory = EconomicMemory()
    return _economic_memory
