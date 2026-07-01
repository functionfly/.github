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


class ThompsonSamplingRouter:
    """Thompson Sampling multi-armed bandit for edge routing.

    Each function maintains its own set of arm states, allowing
    per-function learning of the optimal edge provider.
    """

    ARMS_KEY_PREFIX = "ml:thompson:arms:"
    OUTCOMES_KEY_PREFIX = "ml:thompson:outcomes:"
    ARMS_EXPIRY_DAYS = 30

    def __init__(self):
        self._redis: Optional[redis.Redis] = None
        self._exploration_rate = settings.ml_routing_exploration
        self._arm_cache: Dict[str, Dict[str, ArmState]] = {}

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

    async def _get_arms(self, function_id: str) -> Dict[str, ArmState]:
        """Get arm states for a function."""
        if function_id in self._arm_cache:
            return self._arm_cache[function_id]

        r = await self.get_redis()
        if r:
            try:
                key = f"{self.ARMS_KEY_PREFIX}{function_id}"
                data = await r.get(key)
                if data:
                    arms_data = json.loads(data)
                    arms = {k: ArmState(**v) for k, v in arms_data.items()}
                    self._arm_cache[function_id] = arms
                    return arms
            except Exception as e:
                logger.warning(f"Failed to load arms for {function_id}: {e}")

        # Initialize with equal priors
        arms = {edge: ArmState(edge=edge) for edge in EDGES}
        self._arm_cache[function_id] = arms
        return arms

    async def _save_arms(self, function_id: str, arms: Dict[str, ArmState]) -> None:
        """Persist arm states to Redis."""
        self._arm_cache[function_id] = arms
        r = await self.get_redis()
        if r:
            try:
                key = f"{self.ARMS_KEY_PREFIX}{function_id}"
                data = {k: v.model_dump() for k, v in arms.items()}
                await r.set(key, json.dumps(data), ex=self.ARMS_EXPIRY_DAYS * 86400)
            except Exception as e:
                logger.error(f"Failed to save arms for {function_id}: {e}")

    async def decide(
        self, request: RoutingDecisionRequest
    ) -> RoutingDecision:
        """Make a routing decision using Thompson Sampling.

        For each arm, sample from Beta(alpha, beta) and pick the highest.
        With probability exploration_rate, force exploration of a random arm.
        """
        function_id = request.function_id
        arms = await self._get_arms(function_id)

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

    async def update(self, outcome: RoutingOutcome) -> None:
        """Update arm state based on execution outcome.

        Reward = 0.4 * (1 - normalize(latency)) + 0.4 * success + 0.2 * (1 - normalize(cost))
        """
        arms = await self._get_arms(outcome.function_id)
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
        await self._save_arms(outcome.function_id, arms)

        # Log outcome for analysis
        r = await self.get_redis()
        if r:
            try:
                key = f"{self.OUTCOMES_KEY_PREFIX}{outcome.function_id}"
                await r.lpush(key, outcome.model_dump_json())
                await r.ltrim(key, 0, 999)
                await r.expire(key, 86400 * 7)
            except Exception:
                pass

    async def get_arm_stats(self, function_id: str) -> Dict[str, ArmState]:
        """Get current arm states for a function."""
        return await self._get_arms(function_id)

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
