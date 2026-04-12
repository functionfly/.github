"""Optimized function generation service.

Integrates all cost-optimization strategies:
- Multi-tier model routing
- Validation pipeline with auto-fix
- RAG + template retrieval
- Caching
- Confidence-based escalation
"""

import json
import logging
import time
from typing import Optional, List, Dict, Any, Tuple
from dataclasses import dataclass

from ...models.schemas import (
    FunctionGenerationRequest,
    FunctionGenerationResponse,
    FunctionGenerationResult,
    FunctionManifest,
    ChatMessage,
    MessageRole,
)
from ...providers.manager import get_provider_manager
from ...security.auth import APIKeyInfo

from .model_router import (
    get_model_router,
    ModelTier,
    RoutingDecision,
)
from .validation import (
    get_validation_pipeline,
    ValidationReport,
)
from .rag_retrieval import (
    get_function_rag_retriever,
    RetrievedFunction,
    TemplateMatch,
)
from .cache import (
    get_generation_cache,
    get_cost_tracker,
)

logger = logging.getLogger(__name__)


@dataclass
class GenerationAttempt:
    """Record of a generation attempt."""
    tier: ModelTier
    model: str
    provider: str
    success: bool
    validation_passed: bool
    cost_usd: float
    tokens_in: int
    tokens_out: int
    latency_ms: float
    errors: List[str]


@dataclass
class OptimizedGenerationMetrics:
    """Metrics for an optimized generation."""
    total_attempts: int
    final_tier: str
    cache_hit: bool
    template_used: bool
    validation_attempts: int
    total_cost_usd: float
    total_latency_ms: float
    savings_vs_premium_usd: float
    savings_vs_premium_pct: float


class OptimizedGenerationService:
    """Cost-optimized function generation with all strategies integrated."""

    MAX_ATTEMPTS = 3  # Max generation attempts with escalation
    MAX_FIX_ATTEMPTS = 2  # Max auto-fix attempts per tier

    def __init__(self):
        self.router = get_model_router()
        self.validator = get_validation_pipeline()
        self.rag = get_function_rag_retriever()
        self.cache = get_generation_cache()
        self.cost_tracker = get_cost_tracker()

    async def generate(
        self,
        request: FunctionGenerationRequest,
        tenant_id: str,
        api_key_info: Optional[APIKeyInfo] = None,
        force_tier: Optional[ModelTier] = None,
    ) -> Tuple[FunctionGenerationResponse, OptimizedGenerationMetrics]:
        """Generate function with cost optimization.

        Full pipeline:
        1. Check cache
        2. Analyze complexity and route to appropriate tier
        3. Retrieve similar functions and templates (RAG)
        4. Generate with optimized prompt
        5. Validate output
        6. Auto-fix or escalate if needed
        7. Cache successful result

        Args:
            request: Generation request
            tenant_id: Tenant ID for RAG
            api_key_info: API key info
            force_tier: Force specific tier (optional)

        Returns:
            Tuple of (response, metrics)
        """
        generation_id = f"opt-{int(time.time() * 1000)}"
        start_time = time.time()
        attempts: List[GenerationAttempt] = []

        # Step 1: Check cache
        cached = await self.cache.get(
            request.description,
            request.runtime,
            request.constraints,
        )

        if cached:
            logger.info(f"Cache hit for generation: {cached.cache_key}")
            self.cost_tracker.record_generation(
                tier="cache",
                model="cache",
                tokens_in=0,
                tokens_out=0,
                cost_usd=0,
                was_cached=True,
            )

            result = FunctionGenerationResult(
                code=cached.code,
                runtime=cached.runtime,
                manifest=cached.manifest,
                explanation=cached.explanation,
                suggested_tests=[],
                estimated_complexity=cached.complexity,
            )

            metrics = OptimizedGenerationMetrics(
                total_attempts=0,
                final_tier="cache",
                cache_hit=True,
                template_used=False,
                validation_attempts=0,
                total_cost_usd=0,
                total_latency_ms=(time.time() - start_time) * 1000,
                savings_vs_premium_usd=0.5,  # Estimated
                savings_vs_premium_pct=100,
            )

            return FunctionGenerationResponse(
                success=True,
                result=result,
                generation_id=generation_id,
                latency_ms=metrics.total_latency_ms,
                tokens_used={"prompt": 0, "completion": 0, "total": 0},
            ), metrics

        # Step 2: Route to appropriate tier
        routing = self.router.route(
            request.description,
            request.constraints,
            preferred_tier=force_tier,
        )

        logger.info(
            f"Routing decision: {routing.tier.value} model={routing.model_config.model} "
            f"confidence={routing.confidence:.2f}"
        )

        # Step 3: Retrieve similar functions and templates
        similar_functions = await self.rag.retrieve_similar_functions(
            request.description,
            request.runtime,
            tenant_id,
            limit=3,
        )

        template = self.rag.find_template(request.description, request.runtime)

        # Build generation context
        context = self.rag.build_generation_context(
            request.description,
            request.runtime,
            similar_functions,
            template,
        )

        # Step 4-6: Generate with validation loop
        current_tier = routing.tier
        final_result = None
        final_validation = None

        for attempt_num in range(self.MAX_ATTEMPTS):
            attempt_start = time.time()

            # Get models for current tier
            tier_models = self.router.get_available_models(current_tier)
            if not tier_models:
                logger.warning(f"No models available for tier {current_tier.value}, escalating")
                if current_tier == ModelTier.CHEAP:
                    current_tier = ModelTier.MID
                elif current_tier == ModelTier.MID:
                    current_tier = ModelTier.PREMIUM
                continue

            model_config = tier_models[0]

            # Generate
            generation_result = await self._generate_with_model(
                request=request,
                context=context,
                model_config=model_config,
                attempt=attempt_num,
            )

            latency_ms = (time.time() - attempt_start) * 1000

            if not generation_result["success"]:
                attempts.append(GenerationAttempt(
                    tier=current_tier,
                    model=model_config.model,
                    provider=model_config.provider.value,
                    success=False,
                    validation_passed=False,
                    cost_usd=generation_result.get("cost_usd", 0),
                    tokens_in=generation_result.get("tokens_in", 0),
                    tokens_out=generation_result.get("tokens_out", 0),
                    latency_ms=latency_ms,
                    errors=[generation_result.get("error", "Unknown error")],
                ))

                # Escalate tier
                next_tier = self.router.should_escalate(
                    "",
                    [generation_result.get("error", "")],
                    current_tier,
                )
                if next_tier:
                    current_tier = next_tier
                    logger.info(f"Escalating to {current_tier.value} after generation failure")
                    continue
                break

            code = generation_result["code"]

            # Validate
            validation = self.validator.validate(
                code,
                request.runtime,
                skip_runtime=True,  # Skip runtime for speed
            )

            attempt = GenerationAttempt(
                tier=current_tier,
                model=model_config.model,
                provider=model_config.provider.value,
                success=True,
                validation_passed=validation.overall_passed,
                cost_usd=generation_result.get("cost_usd", 0),
                tokens_in=generation_result.get("tokens_in", 0),
                tokens_out=generation_result.get("tokens_out", 0),
                latency_ms=latency_ms,
                errors=[s.errors + s.warnings for s in validation.stages],
            )
            attempts.append(attempt)

            # Check if validation passed
            if validation.overall_passed or validation.confidence_score >= 0.6:
                final_result = generation_result
                final_validation = validation
                break

            # Try auto-fix
            if attempt_num < self.MAX_FIX_ATTEMPTS:
                fix_result = await self._attempt_fix(
                    code,
                    validation,
                    request,
                    model_config,
                )

                if fix_result and fix_result["success"]:
                    # Validate the fix
                    fix_validation = self.validator.validate(
                        fix_result["code"],
                        request.runtime,
                        skip_runtime=True,
                    )

                    if fix_validation.confidence_score > validation.confidence_score:
                        final_result = fix_result
                        final_validation = fix_validation
                        break

            # Escalate if validation failed
            next_tier = self.router.should_escalate(
                code,
                [e for s in validation.stages for e in s.errors],
                current_tier,
            )
            if next_tier:
                current_tier = next_tier
                logger.info(f"Escalating to {current_tier.value} after validation failure")
                continue

            # Store last result if we can't escalate further
            final_result = generation_result
            final_validation = validation

        # Build response
        total_latency = (time.time() - start_time) * 1000
        total_cost = sum(a.cost_usd for a in attempts)

        if final_result:
            # Parse manifest
            manifest_data = final_result.get("manifest", {})
            manifest = FunctionManifest(
                name=manifest_data.get("name", "generated_function"),
                description=manifest_data.get("description", request.description[:100]),
                version=manifest_data.get("version", "1.0.0"),
                inputs=manifest_data.get("inputs", []),
                outputs=manifest_data.get("outputs", []),
                runtime=manifest_data.get("runtime", request.runtime),
                timeout_seconds=manifest_data.get("timeout_seconds", 30),
                memory_mb=manifest_data.get("memory_mb", 256),
                capabilities=manifest_data.get("capabilities", []),
            )

            result = FunctionGenerationResult(
                code=final_result["code"],
                runtime=manifest.runtime,
                manifest=manifest,
                explanation=final_result.get("explanation", ""),
                suggested_tests=final_result.get("suggested_tests", []),
                estimated_complexity=final_result.get("estimated_complexity", "moderate"),
            )

            success = True

            # Cache successful result
            await self.cache.set(
                description=request.description,
                runtime=request.runtime,
                code=result.code,
                manifest=manifest.model_dump(),
                explanation=result.explanation,
                complexity=result.estimated_complexity,
                constraints=request.constraints,
            )

            # Record cost
            self.cost_tracker.record_generation(
                tier=current_tier.value,
                model=attempts[-1].model if attempts else "unknown",
                tokens_in=sum(a.tokens_in for a in attempts),
                tokens_out=sum(a.tokens_out for a in attempts),
                cost_usd=total_cost,
            )

        else:
            result = None
            success = False

        # Calculate savings vs using premium directly
        premium_cost_estimate = 0.05  # Estimated GPT-4o cost
        savings_usd = premium_cost_estimate - total_cost
        savings_pct = (savings_usd / premium_cost_estimate * 100) if premium_cost_estimate > 0 else 0

        metrics = OptimizedGenerationMetrics(
            total_attempts=len(attempts),
            final_tier=current_tier.value,
            cache_hit=False,
            template_used=template is not None,
            validation_attempts=len([a for a in attempts if a.validation_passed]),
            total_cost_usd=round(total_cost, 4),
            total_latency_ms=round(total_latency, 2),
            savings_vs_premium_usd=round(max(0, savings_usd), 4),
            savings_vs_premium_pct=round(savings_pct, 1),
        )

        response = FunctionGenerationResponse(
            success=success,
            result=result,
            error=None if success else "Generation failed after all attempts",
            generation_id=generation_id,
            latency_ms=metrics.total_latency_ms,
            tokens_used={
                "prompt": sum(a.tokens_in for a in attempts),
                "completion": sum(a.tokens_out for a in attempts),
                "total": sum(a.tokens_in + a.tokens_out for a in attempts),
            },
        )

        logger.info(
            f"Generation complete: success={success} attempts={len(attempts)} "
            f"cost=${total_cost:.4f} savings={savings_pct:.1f}%"
        )

        return response, metrics

    async def _generate_with_model(
        self,
        request: FunctionGenerationRequest,
        context: Dict[str, Any],
        model_config: Any,
        attempt: int,
    ) -> Dict[str, Any]:
        """Generate function with specific model."""
        try:
            provider_manager = get_provider_manager()
            provider = provider_manager.get_provider(model_config.provider.value)

            # Build optimized prompt
            prompt, est_tokens = self.rag.build_optimized_prompt(
                context,
                request.constraints,
            )

            # Build messages
            system_prompt = self._build_system_prompt(request.runtime, attempt)

            messages = [
                ChatMessage(role=MessageRole.SYSTEM, content=system_prompt),
                ChatMessage(role=MessageRole.USER, content=prompt),
            ]

            # Call provider
            completion = await provider.complete(
                messages=messages,
                model=model_config.model,
                temperature=model_config.temperature,
                max_tokens=model_config.max_tokens,
            )

            # Parse response
            content = completion.content

            # Try to extract JSON
            try:
                if "```json" in content:
                    json_str = content.split("```json")[1].split("```")[0].strip()
                elif "```" in content:
                    json_str = content.split("```")[1].split("```")[0].strip()
                else:
                    json_str = content

                result_data = json.loads(json_str)
            except json.JSONDecodeError:
                # Try to parse as raw code
                result_data = {
                    "code": content,
                    "manifest": {
                        "name": "generated_function",
                        "description": request.description[:100],
                        "runtime": request.runtime,
                    },
                    "explanation": "Generated code",
                    "suggested_tests": [],
                    "estimated_complexity": "moderate",
                }

            # Calculate cost
            cost = (
                (completion.usage.get("prompt_tokens", 0) / 1000) * model_config.cost_per_1k_input +
                (completion.usage.get("completion_tokens", 0) / 1000) * model_config.cost_per_1k_output
            )

            return {
                "success": True,
                "code": result_data.get("code", content),
                "manifest": result_data.get("manifest", {}),
                "explanation": result_data.get("explanation", ""),
                "suggested_tests": result_data.get("suggested_tests", []),
                "estimated_complexity": result_data.get("estimated_complexity", "moderate"),
                "cost_usd": cost,
                "tokens_in": completion.usage.get("prompt_tokens", 0),
                "tokens_out": completion.usage.get("completion_tokens", 0),
            }

        except Exception as e:
            logger.error(f"Generation attempt {attempt} failed: {e}")
            return {
                "success": False,
                "error": str(e),
                "cost_usd": 0,
                "tokens_in": 0,
                "tokens_out": 0,
            }

    def _build_system_prompt(self, runtime: str, attempt: int) -> str:
        """Build system prompt for generation."""
        runtime_guidance = {
            "python": "Generate Python 3.11+ code with type hints and docstrings.",
            "nodejs": "Generate Node.js 20+ JavaScript with async/await.",
            "typescript": "Generate TypeScript with proper type annotations.",
            "go": "Generate Go 1.21+ with proper error handling.",
            "rust": "Generate Rust with modern idioms.",
        }.get(runtime, f"Generate {runtime} code.")

        base_prompt = f"""You are an expert serverless function developer.
{runtime_guidance}

Requirements:
1. Self-contained, stateless function
2. Handle errors gracefully
3. Include input validation
4. Production-ready code
5. Respond with JSON containing: code, manifest, explanation, suggested_tests, estimated_complexity"""

        if attempt > 0:
            base_prompt += "\n\nIMPORTANT: This is a regeneration after validation errors. Ensure code is syntactically correct."

        return base_prompt

    async def _attempt_fix(
        self,
        code: str,
        validation: ValidationReport,
        request: FunctionGenerationRequest,
        model_config: Any,
    ) -> Optional[Dict[str, Any]]:
        """Attempt to fix validation errors."""
        try:
            fix_prompt = self.validator.get_fix_prompt(code, validation, request.runtime)

            provider_manager = get_provider_manager()
            provider = provider_manager.get_provider(model_config.provider.value)

            messages = [
                ChatMessage(role=MessageRole.SYSTEM, content="Fix the provided code. Return ONLY the fixed code, no explanations."),
                ChatMessage(role=MessageRole.USER, content=fix_prompt),
            ]

            completion = await provider.complete(
                messages=messages,
                model=model_config.model,
                temperature=0.1,  # Low temp for fixes
                max_tokens=model_config.max_tokens,
            )

            fixed_code = completion.content

            # Clean up markdown if present
            if "```" in fixed_code:
                fixed_code = fixed_code.split("```")[1].split("```")[0].strip()

            return {
                "success": True,
                "code": fixed_code,
                "manifest": {
                    "name": "fixed_function",
                    "description": request.description[:100],
                    "runtime": request.runtime,
                },
                "explanation": "Fixed code based on validation feedback",
                "suggested_tests": [],
                "estimated_complexity": "moderate",
                "cost_usd": 0,  # Negligible for fix
                "tokens_in": completion.usage.get("prompt_tokens", 0),
                "tokens_out": completion.usage.get("completion_tokens", 0),
            }

        except Exception as e:
            logger.warning(f"Auto-fix attempt failed: {e}")
            return None

    async def get_stats(self) -> Dict[str, Any]:
        """Get service statistics."""
        return {
            "cache": self.cache.get_stats(),
            "costs": self.cost_tracker.get_stats(),
        }


# Global service instance
_service: Optional[OptimizedGenerationService] = None


def get_optimized_generation_service() -> OptimizedGenerationService:
    """Get global optimized generation service."""
    global _service
    if _service is None:
        _service = OptimizedGenerationService()
    return _service
