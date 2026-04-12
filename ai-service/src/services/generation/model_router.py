"""Multi-tier model routing for cost-optimized AI generation.

Routes requests to appropriate model tiers based on complexity estimation:
- Tier 1 (Cheap): 80% of simple functions - Ollama/local models
- Tier 2 (Mid): 15% of moderate complexity - OpenRouter budget models
- Tier 3 (Premium): 5% of complex functions - GPT-4o/Claude
"""

import re
import logging
from typing import Optional, Dict, List, Tuple
from dataclasses import dataclass
from enum import Enum

from ...config import settings
from ...models.schemas import ProviderType

logger = logging.getLogger(__name__)


class ModelTier(Enum):
    """Model tiers for cost optimization."""
    CHEAP = "cheap"      # Local/budget models
    MID = "mid"          # Mid-range API models
    PREMIUM = "premium"  # Top-tier models


@dataclass
class ModelConfig:
    """Configuration for a model tier."""
    provider: ProviderType
    model: str
    max_tokens: int
    temperature: float
    cost_per_1k_input: float   # USD per 1K tokens
    cost_per_1k_output: float  # USD per 1K tokens
    avg_latency_ms: int


# Tier configurations
TIER_MODELS = {
    ModelTier.CHEAP: [
        ModelConfig(
            provider=ProviderType.OLLAMA,
            model="qwen2.5-coder:14b",
            max_tokens=2000,
            temperature=0.2,
            cost_per_1k_input=0.0,
            cost_per_1k_output=0.0,
            avg_latency_ms=800,
        ),
        ModelConfig(
            provider=ProviderType.OLLAMA,
            model="codellama:13b",
            max_tokens=2000,
            temperature=0.2,
            cost_per_1k_input=0.0,
            cost_per_1k_output=0.0,
            avg_latency_ms=600,
        ),
    ],
    ModelTier.MID: [
        ModelConfig(
            provider=ProviderType.OPENROUTER,
            model="google/gemini-flash-1.5",
            max_tokens=4000,
            temperature=0.2,
            cost_per_1k_input=0.000075,
            cost_per_1k_output=0.0003,
            avg_latency_ms=400,
        ),
        ModelConfig(
            provider=ProviderType.OPENAI,
            model="gpt-4o-mini",
            max_tokens=4000,
            temperature=0.2,
            cost_per_1k_input=0.00015,
            cost_per_1k_output=0.0006,
            avg_latency_ms=300,
        ),
    ],
    ModelTier.PREMIUM: [
        ModelConfig(
            provider=ProviderType.OPENAI,
            model="gpt-4o",
            max_tokens=4000,
            temperature=0.2,
            cost_per_1k_input=0.0025,
            cost_per_1k_output=0.01,
            avg_latency_ms=600,
        ),
        ModelConfig(
            provider=ProviderType.ANTHROPIC,
            model="claude-3-5-sonnet-20241022",
            max_tokens=4000,
            temperature=0.2,
            cost_per_1k_input=0.003,
            cost_per_1k_output=0.015,
            avg_latency_ms=700,
        ),
    ],
}


class ComplexityAnalyzer:
    """Analyzes function generation request complexity."""

    # Complexity keywords
    SIMPLE_KEYWORDS = [
        "summarize", "parse", "validate", "format", "convert",
        "simple", "basic", "string", "number", "json", "email",
        "webhook", "notify", "log", "filter", "sort", "count",
        "hello", "greeting", "echo", "ping", "health", "status"
    ]

    COMPLEX_KEYWORDS = [
        "workflow", "pipeline", "orchestrate", "multi-step",
        "machine learning", "ai", "llm", "embeddings", "vector",
        "optimization", "algorithm", "cryptography", "secure",
        "distributed", "consensus", "blockchain", "real-time",
        "streaming", "websocket", "complex", "advanced", "sophisticated"
    ]

    MODERATE_KEYWORDS = [
        "api", "database", "db", "auth", "authentication",
        "cache", "queue", "storage", "http", "rest", "graphql",
        "integration", "third-party", "external", "service",
        "process", "transform", "aggregate", "analyze"
    ]

    @classmethod
    def analyze(cls, description: str, constraints: Optional[str] = None) -> Tuple[ModelTier, float]:
        """Analyze complexity and return appropriate tier with confidence.

        Returns:
            Tuple of (ModelTier, confidence_score)
        """
        text = (description + " " + (constraints or "")).lower()

        # Count keyword matches
        simple_score = sum(1 for kw in cls.SIMPLE_KEYWORDS if kw in text)
        complex_score = sum(1 for kw in cls.COMPLEX_KEYWORDS if kw in text)
        moderate_score = sum(1 for kw in cls.MODERATE_KEYWORDS if kw in text)

        # Check for explicit complexity indicators
        if "very complex" in text or "highly sophisticated" in text:
            complex_score += 3
        if "simple" in text or "basic" in text:
            simple_score += 2

        # Check for technical indicators
        indicators = {
            "api_calls": len(re.findall(r'\b(api|endpoint|http|rest|graphql)\b', text)),
            "db_ops": len(re.findall(r'\b(database|db|sql|query|store|save)\b', text)),
            "auth": len(re.findall(r'\b(auth|login|token|jwt|password|permission)\b', text)),
            "async": len(re.findall(r'\b(async|await|callback|promise|concurrent|parallel)\b', text)),
            "steps": len(re.findall(r'\b(step|stage|phase|then|after|before|first|next|finally)\b', text)),
        }

        # Calculate complexity score
        score = 0
        score += simple_score * -1  # Simple reduces score
        score += complex_score * 2   # Complex increases score more
        score += moderate_score * 0.5
        score += indicators["api_calls"] * 0.5
        score += indicators["db_ops"] * 0.5
        score += indicators["auth"] * 1
        score += indicators["async"] * 0.3
        score += indicators["steps"] * 0.8

        # Length factor (longer descriptions tend to be more complex)
        word_count = len(text.split())
        if word_count > 100:
            score += 1
        if word_count > 200:
            score += 1.5

        # Determine tier based on score
        if score >= 3 or complex_score >= 2:
            return ModelTier.PREMIUM, min(0.95, 0.7 + complex_score * 0.1)
        elif score <= -2 or (simple_score >= 2 and complex_score == 0):
            return ModelTier.CHEAP, min(0.95, 0.8 + simple_score * 0.05)
        else:
            return ModelTier.MID, 0.75


@dataclass
class RoutingDecision:
    """Routing decision with model selection."""
    tier: ModelTier
    model_config: ModelConfig
    confidence: float
    estimated_cost_usd: float
    reasoning: str


class ModelRouter:
    """Intelligent model router for cost-optimized generation."""

    def __init__(self):
        self.analyzer = ComplexityAnalyzer()
        self._provider_availability: Dict[ProviderType, bool] = {}

    def update_provider_availability(self, availability: Dict[ProviderType, bool]) -> None:
        """Update which providers are available."""
        self._provider_availability = availability

    def get_available_models(self, tier: ModelTier) -> List[ModelConfig]:
        """Get available models for a tier."""
        models = TIER_MODELS.get(tier, [])
        return [
            m for m in models
            if self._provider_availability.get(m.provider, True)
        ]

    def route(
        self,
        description: str,
        constraints: Optional[str] = None,
        preferred_tier: Optional[ModelTier] = None,
    ) -> RoutingDecision:
        """Route to appropriate model based on complexity.

        Args:
            description: Function description
            constraints: Optional constraints
            preferred_tier: Force a specific tier (optional)

        Returns:
            RoutingDecision with selected model
        """
        # Analyze complexity
        tier, confidence = self.analyzer.analyze(description, constraints)

        # Override if preferred tier specified
        if preferred_tier:
            tier = preferred_tier
            confidence = 0.9

        # Get available models for tier
        available = self.get_available_models(tier)

        # Fallback to higher tier if no models available
        if not available:
            if tier == ModelTier.CHEAP:
                logger.warning("No cheap models available, escalating to mid tier")
                tier = ModelTier.MID
                available = self.get_available_models(tier)
            if not available and tier == ModelTier.MID:
                logger.warning("No mid models available, escalating to premium")
                tier = ModelTier.PREMIUM
                available = self.get_available_models(tier)

        if not available:
            raise RuntimeError("No models available across all tiers")

        # Select first available model (could add load balancing here)
        selected = available[0]

        # Estimate cost (rough calculation)
        est_input_tokens = len(description.split()) * 2 + 500  # System prompt
        est_output_tokens = selected.max_tokens * 0.5  # Assume 50% utilization

        est_cost = (
            (est_input_tokens / 1000) * selected.cost_per_1k_input +
            (est_output_tokens / 1000) * selected.cost_per_1k_output
        )

        reasoning = self._generate_reasoning(tier, confidence, selected)

        return RoutingDecision(
            tier=tier,
            model_config=selected,
            confidence=confidence,
            estimated_cost_usd=est_cost,
            reasoning=reasoning,
        )

    def _generate_reasoning(self, tier: ModelTier, confidence: float, model: ModelConfig) -> str:
        """Generate human-readable reasoning."""
        tier_names = {
            ModelTier.CHEAP: "local/budget",
            ModelTier.MID: "mid-range",
            ModelTier.PREMIUM: "premium",
        }
        return (
            f"Selected {tier_names[tier]} model {model.model} "
            f"from {model.provider.value} with {confidence:.0%} confidence. "
            f"Estimated cost: ${model.estimated_cost_usd:.4f}"
        )

    def should_escalate(
        self,
        result: str,
        validation_errors: List[str],
        current_tier: ModelTier,
    ) -> Optional[ModelTier]:
        """Determine if we should escalate to a higher tier.

        Args:
            result: Generated code
            validation_errors: List of validation errors
            current_tier: Current model tier

        Returns:
            Next tier to try, or None if no escalation needed
        """
        # Don't escalate from premium
        if current_tier == ModelTier.PREMIUM:
            return None

        # Escalate if syntax errors detected
        syntax_error_patterns = [
            "syntax error", "unexpected token", "invalid syntax",
            "parse error", "indentation error", "unexpected indent"
        ]
        has_syntax_error = any(
            p in e.lower() for e in validation_errors for p in syntax_error_patterns
        )

        # Escalate if empty or very short result
        if len(result.strip()) < 100:
            has_syntax_error = True

        # Escalate if multiple validation errors
        if len(validation_errors) >= 3:
            has_syntax_error = True

        if has_syntax_error:
            if current_tier == ModelTier.CHEAP:
                return ModelTier.MID
            if current_tier == ModelTier.MID:
                return ModelTier.PREMIUM

        return None


# Global router instance
_router: Optional[ModelRouter] = None


def get_model_router() -> ModelRouter:
    """Get global model router instance."""
    global _router
    if _router is None:
        _router = ModelRouter()
    return _router
