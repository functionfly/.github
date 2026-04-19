"""Economic routing service for cost-intelligent provider selection.

Extends the base routing service with cost-per-quality metrics to make
intelligent routing decisions that balance cost and quality.
"""

import logging
from dataclasses import dataclass
from typing import Dict, List, Optional, Tuple
from enum import Enum

from ...config import settings
from ...models.schemas import ProviderType, RoutingDecision, RoutingDecisionRequest, EdgeProvider
from ..routing.selector import RoutingService, get_routing_service
from ..economic_memory import CostQualityScore, get_economic_memory

logger = logging.getLogger(__name__)


class RoutingStrategy(str, Enum):
    """Available routing strategies."""
    QUALITY_FIRST = "quality_first"  # Maximize quality regardless of cost
    BALANCED = "balanced"  # Balance cost and quality (default)
    COST_OPTIMIZED = "cost_optimized"  # Minimize cost while meeting quality threshold
    COST_FIRST = "cost_first"  # Minimize cost (may reduce quality)


@dataclass
class EconomicRoutingScore:
    """Enhanced routing score with economic factors."""
    provider: ProviderType
    model: str
    
    # Base routing scores (from base RoutingService)
    latency_score: float = 0.0
    load_score: float = 0.0
    availability_score: float = 0.0
    
    # Economic scores
    cost_per_1k_tokens: float = 0.0
    cost_quality_index: float = 0.0
    quality_score: float = 0.0
    success_rate: float = 1.0
    
    # Combined economic score
    economic_score: float = 0.0
    
    # Final score (weighted combination)
    final_score: float = 0.0
    
    # Metadata
    confidence: str = "high"  # high, medium, low (based on data volume)
    data_points: int = 0


class EconomicRoutingService:
    """Routing service that considers cost-per-quality metrics.
    
    This service extends the base routing with economic intelligence:
    - Tracks which provider/model combinations offer best value
    - Makes routing decisions based on cost-quality balance
    - Suggests model downgrades when quality permits
    - Recommends upgrades when quality suffers
    """
    
    # Provider cost estimates (per 1K tokens) - baseline when no data
    BASELINE_COSTS = {
        ProviderType.OPENAI: 0.002,  # GPT-3.5 turbo
        ProviderType.ANTHROPIC: 0.003,  # Claude Haiku
        ProviderType.GROQ: 0.0001,  # Very cheap
        ProviderType.TOGETHER: 0.0002,
        ProviderType.DEEPINFRA: 0.00015,
        ProviderType.FIREWORKS: 0.0002,
        ProviderType.OPENROUTER: 0.0005,
        ProviderType.OLLAMA: 0.0,  # Self-hosted
    }
    
    # Quality tiers (baseline when no data)
    BASELINE_QUALITY = {
        ProviderType.OPENAI: 0.85,
        ProviderType.ANTHROPIC: 0.88,
        ProviderType.GROQ: 0.75,
        ProviderType.TOGETHER: 0.78,
        ProviderType.DEEPINFRA: 0.76,
        ProviderType.FIREWORKS: 0.77,
        ProviderType.OPENROUTER: 0.80,
        ProviderType.OLLAMA: 0.70,
    }
    
    # Model quality mappings
    MODEL_QUALITY_OVERRIDES = {
        "gpt-4o": 0.95,
        "gpt-4o-mini": 0.82,
        "gpt-3.5-turbo": 0.75,
        "claude-3-opus": 0.95,
        "claude-3-sonnet": 0.88,
        "claude-3-haiku": 0.80,
        "llama-3.1-70b": 0.85,
        "llama-3.1-8b": 0.75,
        "mixtral-8x22b": 0.82,
    }
    
    # Routing weights
    DEFAULT_WEIGHTS = {
        "latency": 0.25,
        "load": 0.20,
        "availability": 0.15,
        "economic": 0.40,  # Cost-quality combined
    }
    
    def __init__(self):
        self._base_router = get_routing_service()
        self._economic_memory = get_economic_memory()
        self._weights = self.DEFAULT_WEIGHTS.copy()
    
    async def decide_routing(
        self,
        request: RoutingDecisionRequest,
        strategy: RoutingStrategy = RoutingStrategy.BALANCED,
        quality_threshold: Optional[float] = None,
        max_cost_per_1k: Optional[float] = None,
    ) -> RoutingDecision:
        """Determine optimal provider with economic considerations.
        
        Args:
            request: Routing decision request
            strategy: Routing strategy to use
            quality_threshold: Minimum quality score (0-1) required
            max_cost_per_1k: Maximum cost per 1K tokens allowed
            
        Returns:
            RoutingDecision with economic insights
        """
        # Get base routing decision for edge selection
        base_decision = await self._base_router.decide_routing(request)
        
        # Get provider scores with economic factors
        provider_scores = await self._score_providers(
            strategy=strategy,
            quality_threshold=quality_threshold,
            max_cost_per_1k=max_cost_per_1k,
        )
        
        if not provider_scores:
            # Fall back to base decision with defaults
            return base_decision
        
        # Select best provider based on strategy
        best = self._select_best_provider(provider_scores, strategy)
        
        if not best:
            return base_decision
        
        # Build economic reasoning
        reasoning_parts = [
            base_decision.reasoning,
            f"Selected {best.provider.value}/{best.model} with",
            f"cost-quality index={best.cost_quality_index:.1f},",
            f"quality={best.quality_score:.2f},",
            f"cost=${best.cost_per_1k_tokens:.4f}/1k tokens",
        ]
        
        # Add strategy-specific reasoning
        if strategy == RoutingStrategy.COST_OPTIMIZED:
            reasoning_parts.append(f"(cost-optimized, min_quality={quality_threshold or 0.7})")
        elif strategy == RoutingStrategy.QUALITY_FIRST:
            reasoning_parts.append("(quality-first)")
        elif strategy == RoutingStrategy.COST_FIRST:
            reasoning_parts.append("(cost-first)")
        else:
            reasoning_parts.append("(balanced)")
        
        return RoutingDecision(
            function_id=request.function_id,
            recommended_edge=base_decision.recommended_edge,  # Keep edge recommendation
            confidence=best.final_score,
            reasoning=" ".join(reasoning_parts),
            alternatives=base_decision.alternatives,
            latency_estimate_ms=base_decision.latency_estimate_ms,
            decided_at=base_decision.decided_at,
        )
    
    async def _score_providers(
        self,
        strategy: RoutingStrategy,
        quality_threshold: Optional[float],
        max_cost_per_1k: Optional[float],
    ) -> List[EconomicRoutingScore]:
        """Score all providers with economic factors."""
        scores = []
        
        # Get economic memory scores
        memory_scores = await self._economic_memory.get_all_scores()
        
        # Build lookup by provider
        memory_by_provider: Dict[Tuple[ProviderType, str], CostQualityScore] = {}
        for score in memory_scores:
            memory_by_provider[(score.provider, score.model)] = score
        
        # Score each provider-model combination
        for provider in ProviderType:
            # Get default model for provider
            model = self._get_default_model(provider)
            
            score = await self._score_provider(
                provider,
                model,
                memory_by_provider.get((provider, model)),
            )
            
            # Apply strategy filters
            if quality_threshold and score.quality_score < quality_threshold:
                continue
            if max_cost_per_1k and score.cost_per_1k_tokens > max_cost_per_1k:
                continue
            
            # Calculate final score based on strategy
            score.final_score = self._calculate_final_score(score, strategy)
            scores.append(score)
        
        return scores
    
    async def _score_provider(
        self,
        provider: ProviderType,
        model: str,
        memory_score: Optional[CostQualityScore],
    ) -> EconomicRoutingScore:
        """Score a single provider with economic factors."""
        score = EconomicRoutingScore(
            provider=provider,
            model=model,
        )
        
        if memory_score and memory_score.total_executions >= 5:
            # Use measured data
            score.cost_per_1k_tokens = memory_score.avg_cost_per_1k_tokens
            score.cost_quality_index = memory_score.cost_quality_index
            score.quality_score = memory_score.quality_score
            score.success_rate = memory_score.success_rate
            score.data_points = memory_score.total_executions
            score.confidence = "high" if memory_score.total_executions >= 50 else "medium"
        else:
            # Use baseline estimates
            score.cost_per_1k_tokens = self.BASELINE_COSTS.get(provider, 0.001)
            score.quality_score = self.MODEL_QUALITY_OVERRIDES.get(
                model, 
                self.BASELINE_QUALITY.get(provider, 0.75)
            )
            score.success_rate = 0.95
            score.confidence = "low"
            
            # Estimate CQI from baselines
            if score.cost_per_1k_tokens > 0:
                score.cost_quality_index = (score.quality_score * 100) / (score.cost_per_1k_tokens * 1000)
            else:
                score.cost_quality_index = score.quality_score * 100
        
        # Calculate economic score component (0-1 scale)
        # Higher CQI = better economic value
        score.economic_score = min(1.0, score.cost_quality_index / 50.0)
        
        return score
    
    def _calculate_final_score(
        self,
        score: EconomicRoutingScore,
        strategy: RoutingStrategy,
    ) -> float:
        """Calculate final routing score based on strategy."""
        if strategy == RoutingStrategy.QUALITY_FIRST:
            # 60% quality, 20% success rate, 20% economic
            return (
                score.quality_score * 0.6 +
                score.success_rate * 0.2 +
                score.economic_score * 0.2
            )
        
        elif strategy == RoutingStrategy.COST_FIRST:
            # 70% economic, 15% success rate, 15% quality (minimum threshold)
            return (
                score.economic_score * 0.7 +
                score.success_rate * 0.15 +
                max(0.5, score.quality_score) * 0.15
            )
        
        elif strategy == RoutingStrategy.COST_OPTIMIZED:
            # Quality must be above threshold, then optimize for cost
            if score.quality_score < 0.7:
                return score.quality_score * 0.3  # Penalize low quality
            return (
                score.economic_score * 0.5 +
                score.quality_score * 0.3 +
                score.success_rate * 0.2
            )
        
        else:  # BALANCED (default)
            # Equal weight to quality and economics
            return (
                score.quality_score * 0.3 +
                score.economic_score * 0.3 +
                score.success_rate * 0.2 +
                score.latency_score * 0.2
            )
    
    def _select_best_provider(
        self,
        scores: List[EconomicRoutingScore],
        strategy: RoutingStrategy,
    ) -> Optional[EconomicRoutingScore]:
        """Select the best provider from scored options."""
        if not scores:
            return None
        
        # Sort by final score descending
        sorted_scores = sorted(scores, key=lambda s: s.final_score, reverse=True)
        
        # Log selection for analysis
        best = sorted_scores[0]
        logger.info(
            f"Economic routing selected {best.provider.value}/{best.model} "
            f"with score={best.final_score:.3f}, CQI={best.cost_quality_index:.1f} "
            f"(strategy={strategy.value})"
        )
        
        return best
    
    def _get_default_model(self, provider: ProviderType) -> str:
        """Get the default model for a provider."""
        defaults = {
            ProviderType.OPENAI: "gpt-4o-mini",
            ProviderType.ANTHROPIC: "claude-3-haiku",
            ProviderType.GROQ: "llama-3.1-8b",
            ProviderType.TOGETHER: "llama-3.1-8b",
            ProviderType.DEEPINFRA: "llama-3.1-8b",
            ProviderType.FIREWORKS: "llama-3.1-8b",
            ProviderType.OPENROUTER: "openai/gpt-3.5-turbo",
            ProviderType.OLLAMA: "llama3.1",
        }
        return defaults.get(provider, "default")
    
    async def get_model_recommendation(
        self,
        provider: ProviderType,
        current_model: str,
        target_quality: float = 0.75,
        max_cost_increase_pct: float = 20.0,
    ) -> Dict:
        """Get model recommendation for a provider with economic analysis.
        
        Returns:
            Dict with recommendation details
        """
        memory = self._economic_memory
        
        # Get current model score
        current_score = await memory.get_score(provider, current_model)
        
        # Get all scores for this provider
        all_scores = await memory.get_all_scores()
        provider_scores = [
            s for s in all_scores 
            if s.provider == provider and s.model != current_model
        ]
        
        if not provider_scores:
            return {
                "recommendation": "no_data",
                "current_model": current_model,
                "message": "Insufficient data for recommendations",
            }
        
        # Find alternatives with similar quality
        alternatives = []
        for score in provider_scores:
            if score.total_executions < 10:
                continue
            if score.quality_score < target_quality * 0.9:
                continue
            
            alternatives.append(score)
        
        if not alternatives:
            return {
                "recommendation": "keep_current",
                "current_model": current_model,
                "current_quality": current_score.quality_score if current_score else None,
                "message": "No suitable alternatives found",
            }
        
        # Sort by cost-quality index
        alternatives.sort(key=lambda s: s.cost_quality_index, reverse=True)
        best_alt = alternatives[0]
        
        if not current_score:
            return {
                "recommendation": "consider_upgrade",
                "current_model": current_model,
                "suggested_model": best_alt.model,
                "savings_percent": None,
                "message": f"Consider {best_alt.model} for better value",
            }
        
        # Calculate cost difference
        cost_diff = current_score.avg_cost_per_1k_tokens - best_alt.avg_cost_per_1k_tokens
        savings_pct = (cost_diff / current_score.avg_cost_per_1k_tokens) * 100 if current_score.avg_cost_per_1k_tokens > 0 else 0
        
        if savings_pct > 10:
            return {
                "recommendation": "downgrade_suggested",
                "current_model": current_model,
                "current_cost_per_1k": current_score.avg_cost_per_1k_tokens,
                "current_quality": current_score.quality_score,
                "current_cqi": current_score.cost_quality_index,
                "suggested_model": best_alt.model,
                "suggested_cost_per_1k": best_alt.avg_cost_per_1k_tokens,
                "suggested_quality": best_alt.quality_score,
                "suggested_cqi": best_alt.cost_quality_index,
                "potential_savings_percent": round(savings_pct, 1),
                "quality_delta": round(best_alt.quality_score - current_score.quality_score, 2),
                "message": f"Switch to {best_alt.model} to save {savings_pct:.0f}% with similar quality",
            }
        elif savings_pct < -max_cost_increase_pct:
            return {
                "recommendation": "upgrade_suggested",
                "current_model": current_model,
                "suggested_model": best_alt.model,
                "quality_delta": round(best_alt.quality_score - current_score.quality_score, 2),
                "message": f"Consider {best_alt.model} for better quality",
            }
        
        return {
            "recommendation": "keep_current",
            "current_model": current_model,
            "current_cqi": current_score.cost_quality_index,
            "best_alternative_cqi": best_alt.cost_quality_index,
            "message": "Current model is optimal for your needs",
        }
    
    async def get_cost_savings_opportunity(
        self,
        tenant_id: Optional[str] = None,
        days: int = 7,
    ) -> Dict:
        """Analyze potential cost savings from better routing.
        
        Returns:
            Dict with savings analysis
        """
        memory = self._economic_memory
        
        # Get recent execution breakdown
        breakdown = await memory.get_cost_breakdown(days=days)
        
        if breakdown["total_executions"] < 10:
            return {
                "period_days": days,
                "analysis": "insufficient_data",
                "message": "Need more execution data for analysis",
            }
        
        # Get best value provider overall
        best_value = await memory.get_best_value_provider(min_executions=10)
        
        if not best_value:
            return {
                "period_days": days,
                "analysis": "no_alternatives",
                "message": "No alternative providers with sufficient data",
            }
        
        # Calculate potential savings
        current_cost = breakdown["total_cost"]
        
        # Estimate savings if all traffic went to best value provider
        # This is optimistic - assumes quality would be maintained
        estimated_savings = current_cost * 0.15  # Conservative 15% estimate
        
        return {
            "period_days": days,
            "analysis": "savings_available",
            "current_period_cost": round(current_cost, 2),
            "executions_analyzed": breakdown["total_executions"],
            "best_value_provider": f"{best_value.provider.value}/{best_value.model}",
            "best_value_cqi": round(best_value.cost_quality_index, 1),
            "estimated_monthly_savings": round(estimated_savings * 4, 2),  # Extrapolate to month
            "optimization_opportunities": [
                f"Switch to {best_value.provider.value}/{best_value.model} for better value",
                "Enable automatic cost-optimized routing",
            ],
        }


# Global instance
_economic_router: Optional[EconomicRoutingService] = None


def get_economic_routing_service() -> EconomicRoutingService:
    """Get the global economic routing service instance."""
    global _economic_router
    if _economic_router is None:
        _economic_router = EconomicRoutingService()
    return _economic_router
