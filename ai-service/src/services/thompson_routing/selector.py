"""Thompson Sampling router — learns optimal edge per function.

Replaces the static weighted scoring (latency 30%, load 30%, availability 40%)
with a Bayesian bandit that naturally balances exploration and exploitation.
"""

import json
import logging
import random
from datetime import datetime
from typing import Dict, List, Optional

import redis.asyncio as redis

from ...config import settings
from ...models.schemas import EdgeProvider, RoutingDecision, RoutingDecisionRequest
from .models import ArmState, RoutingOutcome, ThompsonDecision

logger = logging.getLogger(__name__)

EDGES = [
    EdgeProvider.CLOUDFLARE.value,
    EdgeProvider.VERCEL.value,
    EdgeProvider.FLY.value,
    EdgeProvider.DENO.value,
    EdgeProvider.FUNCTIONFLY.value,
]

FUNCTION_TYPE_PRIORS = {
    "http_trigger": {
        EdgeProvider.CLOUDFLARE.value: ArmState(edge=EdgeProvider.CLOUDFLARE.value, alpha=2.0, beta=1.0),
        EdgeProvider.VERCEL.value: ArmState(edge=EdgeProvider.VERCEL.value, alpha=2.0, beta=1.0),
        EdgeProvider.FLY.value: ArmState(edge=EdgeProvider.FLY.value, alpha=1.5, beta=1.5),
        EdgeProvider.DENO.value: ArmState(edge=EdgeProvider.DENO.value, alpha=1.5, beta=1.5),
        EdgeProvider.FUNCTIONFLY.value: ArmState(edge=EdgeProvider.FUNCTIONFLY.value, alpha=1.5, beta=1.5),
    },
    "scheduled": {
        EdgeProvider.CLOUDFLARE.value: ArmState(edge=EdgeProvider.CLOUDFLARE.value, alpha=1.5, beta=1.5),
        EdgeProvider.VERCEL.value: ArmState(edge=EdgeProvider.VERCEL.value, alpha=1.5, beta=1.5),
        EdgeProvider.FLY.value: ArmState(edge=EdgeProvider.FLY.value, alpha=2.0, beta=1.0),
        EdgeProvider.DENO.value: ArmState(edge=EdgeProvider.DENO.value, alpha=1.5, beta=1.5),
        EdgeProvider.FUNCTIONFLY.value: ArmState(edge=EdgeProvider.FUNCTIONFLY.value, alpha=2.0, beta=1.0),
    },
    "queue_triggered": {
        EdgeProvider.CLOUDFLARE.value: ArmState(edge=EdgeProvider.CLOUDFLARE.value, alpha=1.5, beta=1.5),
        EdgeProvider.VERCEL.value: ArmState(edge=EdgeProvider.VERCEL.value, alpha=1.5, beta=1.5),
        EdgeProvider.FLY.value: ArmState(edge=EdgeProvider.FLY.value, alpha=2.0, beta=1.0),
        EdgeProvider.DENO.value: ArmState(edge=EdgeProvider.DENO.value, alpha=1.5, beta=1.5),
        EdgeProvider.FUNCTIONFLY.value: ArmState(edge=EdgeProvider.FUNCTIONFLY.value, alpha=2.5, beta=0.5),
    },
    "default": {
        EdgeProvider.CLOUDFLARE.value: ArmState(edge=EdgeProvider.CLOUDFLARE.value, alpha=1.0, beta=1.0),
        EdgeProvider.VERCEL.value: ArmState(edge=EdgeProvider.VERCEL.value, alpha=1.0, beta=1.0),
        EdgeProvider.FLY.value: ArmState(edge=EdgeProvider.FLY.value, alpha=1.0, beta=1.0),
        EdgeProvider.DENO.value: ArmState(edge=EdgeProvider.DENO.value, alpha=1.0, beta=1.0),
        EdgeProvider.FUNCTIONFLY.value: ArmState(edge=EdgeProvider.FUNCTIONFLY.value, alpha=1.0, beta=1.0),
    },
}


class ThompsonSamplingRouter:
    """Thompson Sampling multi-armed bandit for edge routing.

    Each function maintains its own set of arm states, allowing
    per-function learning of the optimal edge provider.
    Tenant-isolated: arms are stored per-tenant-per-function.
    Supports function-type-based priors for faster cold-start.
    """

    ARMS_KEY_PREFIX = "ml:thompson:arms:"
    OUTCOMES_KEY_PREFIX = "ml:thompson:outcomes:"
    ARMS_EXPIRY_DAYS = 30

    def __init__(self):
        self._redis: Optional[redis.Redis] = None
        self._exploration_rate = settings.ml_routing_exploration
        self._arm_cache: Dict[str, Dict[str, ArmState]] = {}
        self._use_informed_priors = getattr(
            settings, 'ml_routing_use_informed_priors', True
        )

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

    def _make_arms_key(self, tenant_id: str, function_id: str) -> str:
        """Create tenant-isolated Redis key for arm states."""
        return f"{self.ARMS_KEY_PREFIX}{tenant_id}:{function_id}"

    def _make_outcomes_key(self, tenant_id: str, function_id: str) -> str:
        """Create tenant-isolated Redis key for outcomes."""
        return f"{self.OUTCOMES_KEY_PREFIX}{tenant_id}:{function_id}"

    def _get_function_type(self, function_id: str) -> str:
        """Extract function type from function_id.

        Attempts to determine the function type from naming conventions.
        Falls back to 'default' if type cannot be determined.
        """
        function_id_lower = function_id.lower()

        if any(pattern in function_id_lower for pattern in ['http', 'webhook', 'api', 'rest']):
            return "http_trigger"
        elif any(pattern in function_id_lower for pattern in ['cron', 'scheduled', 'timer', ' recurring']):
            return "scheduled"
        elif any(pattern in function_id_lower for pattern in ['queue', 'worker', 'job', 'task']):
            return "queue_triggered"

        return "default"

    def _get_priors_for_function(self, function_id: str) -> Dict[str, ArmState]:
        """Get informed priors for a function based on its type.

        Args:
            function_id: The function identifier

        Returns:
            Dict of arm states with informed priors
        """
        if not self._use_informed_priors:
            return {edge: ArmState(edge=edge) for edge in EDGES}

        function_type = self._get_function_type(function_id)
        priors = FUNCTION_TYPE_PRIORS.get(function_type, FUNCTION_TYPE_PRIORS["default"])

        logger.debug(
            f"Using {function_type} priors for function {function_id}: "
            f"{[(k, v.alpha, v.beta) for k, v in priors.items()]}"
        )

        return priors

    async def _get_arms(self, tenant_id: str, function_id: str) -> Dict[str, ArmState]:
        """Get arm states for a tenant-function pair.

        Uses informed priors based on function type for cold-start.
        """
        cache_key = f"{tenant_id}:{function_id}"
        if cache_key in self._arm_cache:
            return self._arm_cache[cache_key]

        r = await self.get_redis()
        if r:
            try:
                key = self._make_arms_key(tenant_id, function_id)
                data = await r.get(key)
                if data:
                    arms_data = json.loads(data)
                    arms = {k: ArmState(**v) for k, v in arms_data.items()}
                    self._arm_cache[cache_key] = arms
                    return arms
            except Exception as e:
                logger.warning(f"Failed to load arms for {tenant_id}/{function_id}: {e}")

        arms = self._get_priors_for_function(function_id)
        self._arm_cache[cache_key] = arms
        return arms

    async def _save_arms(self, tenant_id: str, function_id: str, arms: Dict[str, ArmState]) -> None:
        """Persist arm states to Redis with tenant isolation."""
        cache_key = f"{tenant_id}:{function_id}"
        self._arm_cache[cache_key] = arms
        r = await self.get_redis()
        if r:
            try:
                key = self._make_arms_key(tenant_id, function_id)
                data = {k: v.model_dump() for k, v in arms.items()}
                await r.set(key, json.dumps(data), ex=self.ARMS_EXPIRY_DAYS * 86400)
            except Exception as e:
                logger.error(f"Failed to save arms for {tenant_id}/{function_id}: {e}")

    async def decide(
        self, tenant_id: str, request: RoutingDecisionRequest
    ) -> RoutingDecision:
        """Make a routing decision using Thompson Sampling.

        For each arm, sample from Beta(alpha, beta) and pick the highest.
        With probability exploration_rate, force exploration of a random arm.

        Args:
            tenant_id: Tenant ID for isolation
            request: Routing decision request
        """
        function_id = request.function_id
        arms = await self._get_arms(tenant_id, function_id)

        # Force exploration with small probability
        is_exploration = random.random() < self._exploration_rate

        if is_exploration:
            chosen_edge = random.choice(EDGES)
        else:
            # Sample from each arm's Beta distribution
            sampled = {}
            for edge, arm in arms.items():
                sampled[edge] = random.betavariate(arm.alpha, arm.beta)

            chosen_edge = max(sampled, key=sampled.get)

        # Build alternatives (other edges sorted by their mean)
        other_edges = [
            e for e in sorted(arms.keys(), key=lambda e: arms[e].mean, reverse=True)
            if e != chosen_edge
        ]

        chosen_arm = arms.get(chosen_edge, ArmState(edge=chosen_edge))
        confidence = chosen_arm.mean if chosen_arm.total_pulls > 0 else 0.5

        # Estimate latency from historical data
        latency_estimate = chosen_arm.avg_latency_ms if chosen_arm.avg_latency_ms > 0 else 50.0

        # Parse edge provider
        try:
            provider = EdgeProvider(chosen_edge)
        except ValueError:
            provider = EdgeProvider.CLOUDFLARE

        alternatives = []
        for e in other_edges[:3]:
            try:
                alternatives.append(EdgeProvider(e))
            except ValueError:
                continue

        reasoning_parts = [f"Thompson Sampling (exploitation)" if not is_exploration else "Exploration"]
        if chosen_arm.total_pulls > 0:
            reasoning_parts.append(f"{chosen_arm.total_pulls} samples, {chosen_arm.success_rate:.0%} success")

        return RoutingDecision(
            function_id=function_id,
            recommended_edge=provider,
            confidence=round(confidence, 3),
            reasoning=", ".join(reasoning_parts),
            alternatives=alternatives,
            latency_estimate_ms=round(latency_estimate, 2),
        )

    async def update(self, tenant_id: str, outcome: RoutingOutcome) -> None:
        """Update arm state based on execution outcome.

        Reward = 0.4 * (1 - normalize(latency)) + 0.4 * success + 0.2 * (1 - normalize(cost))

        Args:
            tenant_id: Tenant ID for isolation
            outcome: Routing outcome to record
        """
        arms = await self._get_arms(tenant_id, outcome.function_id)
        edge = outcome.edge

        if edge not in arms:
            arms[edge] = ArmState(edge=edge)

        arm = arms[edge]

        # Calculate reward
        latency_reward = max(0, 1.0 - min(outcome.latency_ms / 500.0, 1.0))
        success_reward = 1.0 if outcome.success else 0.0
        cost_reward = max(0, 1.0 - min(outcome.cost_cents / 1.0, 1.0))

        reward = 0.4 * latency_reward + 0.4 * success_reward + 0.2 * cost_reward

        # Update Beta distribution
        arm.alpha += reward
        arm.beta += (1.0 - reward)
        arm.total_pulls += 1
        arm.total_reward += reward

        # Update running average latency
        if arm.total_pulls == 1:
            arm.avg_latency_ms = outcome.latency_ms
        else:
            arm.avg_latency_ms = (
                arm.avg_latency_ms * (arm.total_pulls - 1) + outcome.latency_ms
            ) / arm.total_pulls

        if outcome.success:
            arm.success_count += 1

        arms[edge] = arm
        await self._save_arms(tenant_id, outcome.function_id, arms)

        # Log outcome for analysis
        r = await self.get_redis()
        if r:
            try:
                key = self._make_outcomes_key(tenant_id, outcome.function_id)
                await r.lpush(key, outcome.model_dump_json())
                await r.ltrim(key, 0, 999)
                await r.expire(key, 86400 * 7)
            except Exception:
                pass

    async def get_arm_stats(self, tenant_id: str, function_id: str) -> Dict[str, ArmState]:
        """Get current arm states for a tenant-function pair."""
        return await self._get_arms(tenant_id, function_id)

    async def close(self):
        if self._redis:
            await self._redis.close()
            self._redis = None


_router: Optional[ThompsonSamplingRouter] = None


def get_thompson_router() -> ThompsonSamplingRouter:
    global _router
    if _router is None:
        _router = ThompsonSamplingRouter()
    return _router
