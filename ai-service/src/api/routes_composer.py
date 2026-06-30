"""AI Composer endpoints for function generation."""

import os
import uuid
import json
import time as pytime
import logging
from typing import AsyncGenerator, Optional, List, Dict, Any

from fastapi import APIRouter, HTTPException, Query, status, Depends, Header
from fastapi.responses import StreamingResponse
from pydantic import BaseModel, Field

from ..security.auth import (
    require_api_key_with_scope,
    APIKeyInfo,
    KeyScope,
)
from ..utils.security import sanitize_error_message
from ..config import settings
from ..models.schemas import (
    ChatMessage,
    MessageRole,
    FunctionGenerationRequest,
    FunctionGenerationResponse,
    FunctionManifest,
    FunctionGenerationResult,
)
from ..providers.manager import get_provider_manager

logger = logging.getLogger(__name__)

router = APIRouter()


class OptimizedGenerationRequest(BaseModel):
    """Request for cost-optimized function generation."""

    description: str = Field(
        ..., min_length=10, max_length=2000, description="Natural language description"
    )
    runtime: str = Field(default="python", description="Target runtime")
    inputs: Optional[List[Dict[str, Any]]] = Field(default=None)
    outputs: Optional[List[Dict[str, Any]]] = Field(default=None)
    constraints: Optional[str] = Field(default=None)
    examples: Optional[List[str]] = Field(default=None)
    force_tier: Optional[str] = Field(default=None, description="Force tier: cheap, mid, premium")


class OptimizedGenerationMetricsResponse(BaseModel):
    """Metrics for optimized generation."""

    total_attempts: int
    final_tier: str
    cache_hit: bool
    template_used: bool
    validation_attempts: int
    total_cost_usd: float
    savings_vs_premium_usd: float
    savings_vs_premium_pct: float


class OptimizedGenerationResponse(BaseModel):
    """Response for optimized function generation."""

    success: bool
    result: Optional[FunctionGenerationResult] = None
    error: Optional[str] = None
    generation_id: str
    latency_ms: float
    tokens_used: Dict[str, int]
    metrics: OptimizedGenerationMetricsResponse
    optimization_notes: List[str]


@router.post("/api/composer/generate", response_model=FunctionGenerationResponse)
async def generate_function(
    request: FunctionGenerationRequest,
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.CHAT_WRITE)),
    x_byok_key: Optional[str] = Header(default=None, alias="X-BYOK-Key"),
    x_byok_provider: Optional[str] = Header(default=None, alias="X-BYOK-Provider"),
    x_key_source: str = Header(default="platform", alias="X-Key-Source"),
):
    """Generate a function using AI based on natural language description.

    This endpoint uses LLM to generate complete function code along with
    I/O manifest, test suggestions, and complexity estimation.

    Args:
        request: Function generation request with description and optional constraints
        api_key: Validated API key with chat:write scope
        x_byok_key: Optional BYOK API key from Go proxy
        x_byok_provider: Optional BYOK provider name from Go proxy
        x_key_source: Key source indicator ("byok" or "platform")

    Returns:
        FunctionGenerationResponse with generated code and metadata
    """
    generation_id = str(uuid.uuid4())
    start_time = pytime.time()

    try:
        provider_manager = get_provider_manager()
        byok_key = x_byok_key if x_key_source == "byok" else None
        byok_provider = x_byok_provider if x_key_source == "byok" else None

        if byok_key and byok_provider:
            provider = provider_manager.get_provider_for_request(byok_provider, byok_key)
            provider_name = byok_provider
        else:
            provider_name = "openrouter"
            provider = provider_manager.get_provider(provider_name)

            if not provider or not provider.available:
                provider_name = "openai"
                provider = provider_manager.get_provider(provider_name)

        if not provider or not provider.available:
            raise HTTPException(
                status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                detail="AI generation service not available",
            )

        # Build the prompt for function generation
        runtime_prompts = {
            "python": "Generate Python 3.11+ code. Use type hints, docstrings, and modern Python patterns.",
            "nodejs": "Generate Node.js 20+ JavaScript code. Use async/await and modern patterns.",
            "go": "Generate Go 1.21+ code. Include proper error handling and idiomatic Go patterns.",
            "rust": "Generate Rust code with proper error handling and modern idioms.",
        }

        runtime_guidance = runtime_prompts.get(request.runtime, f"Generate {request.runtime} code.")

        system_prompt = f"""You are an expert serverless function developer.
Your task is to generate complete, production-ready function code based on a natural language description.

{runtime_guidance}

The function should:
1. Be self-contained and stateless (serverless-appropriate)
2. Handle errors gracefully
3. Include input validation
4. Be secure (no code injection vulnerabilities)
5. Follow best practices for the target runtime

Respond with a JSON object containing:
- code: The complete function code
- manifest: Object with name, description, version, inputs (array), outputs (array), runtime, timeout_seconds, memory_mb, capabilities (array)
- explanation: Brief explanation of what the code does
- suggested_tests: Array of test case descriptions
- estimated_complexity: "simple", "moderate", or "complex"

Inputs should be objects with: name, type, description, required (boolean), default (optional)
Outputs should be objects with: name, type, description
Capabilities are strings like "http", "network", "filesystem", etc."""

        user_prompt = f"Generate a {request.runtime} function that: {request.description}"

        if request.constraints:
            user_prompt += f"\n\nAdditional constraints: {request.constraints}"

        if request.examples:
            user_prompt += f"\n\nExample inputs/outputs: {chr(10).join(request.examples)}"

        messages = [
            ChatMessage(role=MessageRole.SYSTEM, content=system_prompt),
            ChatMessage(role=MessageRole.USER, content=user_prompt),
        ]

        model = "gpt-4o" if provider_name == "openai" else "openrouter/free"
        completion = await provider.complete(
            messages=messages,
            model=model,
            temperature=0.2,
            max_tokens=4000,
        )

        try:
            result_data = json.loads(completion.content)
        except json.JSONDecodeError:
            content = completion.content
            if "```json" in content:
                json_str = content.split("```json")[1].split("```")[0].strip()
            elif "```" in content:
                json_str = content.split("```")[1].split("```")[0].strip()
            else:
                raise ValueError("Failed to parse AI response as JSON. The model returned an unexpected format.")
            result_data = json.loads(json_str)

        manifest_data = result_data.get("manifest", {})
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
            code=result_data.get("code", ""),
            runtime=manifest.runtime,
            manifest=manifest,
            explanation=result_data.get("explanation", ""),
            suggested_tests=result_data.get("suggested_tests", []),
            estimated_complexity=result_data.get("estimated_complexity", "moderate"),
        )

        latency_ms = (pytime.time() - start_time) * 1000

        return FunctionGenerationResponse(
            success=True,
            result=result,
            generation_id=generation_id,
            latency_ms=latency_ms,
            tokens_used=completion.usage,
        )

    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Function generation failed: {e}")
        latency_ms = (pytime.time() - start_time) * 1000
        return FunctionGenerationResponse(
            success=False,
            error=str(e),
            generation_id=generation_id,
            latency_ms=latency_ms,
        )


# Internal endpoint with internal API key auth - used by FRG backend
@router.post("/internal/composer/generate", response_model=FunctionGenerationResponse)
async def internal_generate_function(
    request: FunctionGenerationRequest,
    x_internal_key: Optional[str] = Header(None, alias="X-Internal-Key"),
):
    """Internal function generation endpoint for FRG system use.

    This endpoint requires internal API key authentication via X-Internal-Key header
    and is intended for internal service-to-service communication only.
    DO NOT expose this endpoint publicly - it should only be accessible from
    internal networks or through a properly restricted network policy.
    """
    internal_key = settings.internal_api_key or os.environ.get("INTERNAL_API_KEY", "")
    if not internal_key:
        logger.error("Internal endpoint called but INTERNAL_API_KEY not configured")
        raise HTTPException(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
            detail="Internal authentication not configured",
        )

    if not x_internal_key or x_internal_key != internal_key:
        logger.warning(f"Internal endpoint access attempt with invalid key")
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid internal API key",
            headers={"WWW-Authenticate": "ApiKey"},
        )

    generation_id = str(uuid.uuid4())
    start_time = pytime.time()

    try:
        provider_manager = get_provider_manager()
        provider_name = "openrouter"
        provider = provider_manager.get_provider(provider_name)

        if not provider or not provider.available:
            provider_name = "openai"
            provider = provider_manager.get_provider(provider_name)

        if not provider or not provider.available:
            raise HTTPException(
                status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                detail="AI generation service not available",
            )

        # Build the prompt for function generation
        runtime_prompts = {
            "python": "Generate Python 3.11+ code. Use type hints, docstrings, and modern Python patterns.",
            "nodejs": "Generate Node.js 20+ JavaScript code. Use async/await and modern patterns.",
            "go": "Generate Go 1.21+ code. Include proper error handling and idiomatic Go patterns.",
            "rust": "Generate Rust code with proper error handling and modern idioms.",
        }

        runtime_guidance = runtime_prompts.get(request.runtime, f"Generate {request.runtime} code.")

        system_prompt = f"""You are an expert serverless function developer.
Your task is to generate complete, production-ready function code based on a natural language description.

{runtime_guidance}

The function should:
1. Be self-contained and stateless (serverless-appropriate)
2. Handle errors gracefully
3. Include input validation
4. Be secure (no code injection vulnerabilities)
5. Follow best practices for the target runtime

Respond with a JSON object containing:
- code: The complete function code
- manifest: Object with name, description, version, inputs (array), outputs (array), runtime, timeout_seconds, memory_mb, capabilities (array)
- explanation: Brief explanation of what the code does
- suggested_tests: Array of test case descriptions
- estimated_complexity: "simple", "moderate", or "complex"

Inputs should be objects with: name, type, description, required (boolean), default (optional)
Outputs should be objects with: name, type, description
Capabilities are strings like "http", "network", "filesystem", etc."""

        user_prompt = f"Generate a {request.runtime} function that: {request.description}"

        if request.constraints:
            user_prompt += f"\n\nAdditional constraints: {request.constraints}"

        if request.examples:
            user_prompt += f"\n\nExample inputs/outputs: {chr(10).join(request.examples)}"

        messages = [
            ChatMessage(role=MessageRole.SYSTEM, content=system_prompt),
            ChatMessage(role=MessageRole.USER, content=user_prompt),
        ]

        model = "gpt-4o" if provider_name == "openai" else "openrouter/free"
        completion = await provider.complete(
            messages=messages,
            model=model,
            temperature=0.2,
            max_tokens=4000,
        )

        try:
            result_data = json.loads(completion.content)
        except json.JSONDecodeError:
            content = completion.content
            if "```json" in content:
                json_str = content.split("```json")[1].split("```")[0].strip()
            elif "```" in content:
                json_str = content.split("```")[1].split("```")[0].strip()
            else:
                raise ValueError("Failed to parse AI response as JSON")
            result_data = json.loads(json_str)

        manifest_data = result_data.get("manifest", {})
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
            code=result_data.get("code", ""),
            runtime=manifest.runtime,
            manifest=manifest,
            explanation=result_data.get("explanation", ""),
            suggested_tests=result_data.get("suggested_tests", []),
            estimated_complexity=result_data.get("estimated_complexity", "moderate"),
        )

        latency_ms = (pytime.time() - start_time) * 1000

        return FunctionGenerationResponse(
            success=True,
            result=result,
            generation_id=generation_id,
            latency_ms=latency_ms,
            tokens_used=completion.usage,
        )

    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Internal function generation failed: {e}")
        latency_ms = (pytime.time() - start_time) * 1000
        return FunctionGenerationResponse(
            success=False,
            error=str(e),
            generation_id=generation_id,
            latency_ms=latency_ms,
        )


@router.post("/api/composer/generate/stream")
async def generate_function_stream(
    request: FunctionGenerationRequest,
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.CHAT_WRITE)),
    x_byok_key: Optional[str] = Header(default=None, alias="X-BYOK-Key"),
    x_byok_provider: Optional[str] = Header(default=None, alias="X-BYOK-Provider"),
    x_key_source: str = Header(default="platform", alias="X-Key-Source"),
):
    """Stream a function generation using AI.

    This endpoint streams the LLM output as it's generated, useful for
    real-time UI updates during function creation.

    Args:
        request: Function generation request
        api_key: Validated API key with chat:write scope
        x_byok_key: Optional BYOK API key from Go proxy
        x_byok_provider: Optional BYOK provider name from Go proxy
        x_key_source: Key source indicator ("byok" or "platform")

    Returns:
        Streaming response with generated content chunks
    """
    generation_id = str(uuid.uuid4())

    if not settings.enable_streaming:
        raise HTTPException(
            status_code=status.HTTP_501_NOT_IMPLEMENTED,
            detail="Streaming is not enabled",
        )

    try:
        provider_manager = get_provider_manager()
        byok_key = x_byok_key if x_key_source == "byok" else None
        byok_provider = x_byok_provider if x_key_source == "byok" else None

        if byok_key and byok_provider:
            provider = provider_manager.get_provider_for_request(byok_provider, byok_key)
        else:
            provider = provider_manager.get_provider("openai")

        if not provider or not provider.available:
            raise HTTPException(
                status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                detail="AI generation service not available",
            )

        system_prompt = f"""You are an expert serverless function developer.
Generate {request.runtime} code based on the user's description.

Output ONLY the raw code, no markdown formatting, no explanations.
The code should be self-contained and production-ready."""

        messages = [
            ChatMessage(role=MessageRole.SYSTEM, content=system_prompt),
            ChatMessage(role=MessageRole.USER, content=request.description),
        ]

        async def generate() -> AsyncGenerator[str, None]:
            yield f'{{"generation_id": "{generation_id}", "type": "start"}}\n\n'

            try:
                async for chunk in provider.stream(
                    messages=messages,
                    model="gpt-4o",
                    temperature=0.2,
                    max_tokens=4000,
                ):
                    escaped = chunk.replace('\\', '\\\\').replace('"', '\\"').replace('\n', '\\n').replace('\r', '\\r')
                    yield f'{{"type": "chunk", "content": "{escaped}"}}\n\n'

                yield f'{{"type": "complete"}}\n\n'
            except Exception as e:
                error_json = json.dumps({"type": "error", "error": str(e)})
                yield f'{error_json}\n\n'

        return StreamingResponse(
            generate(),
            media_type="text/event-stream",
        )

    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Function generation stream failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to stream function generation",
        )


@router.post("/api/composer/generate-optimized", response_model=OptimizedGenerationResponse)
async def generate_function_optimized(
    request: OptimizedGenerationRequest,
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.CHAT_WRITE)),
    tenant_id: str = Query("default", description="Tenant ID"),
):
    """Generate a function using cost-optimized multi-tier AI pipeline.

    This endpoint implements the complete cost optimization strategy:
    - Multi-tier model routing (cheap -> mid -> premium)
    - Template + RAG retrieval for faster generation
    - Validation pipeline with auto-fix
    - Intelligent caching
    - Confidence-based escalation

    **Cost savings**: 70-90% cheaper than using premium models directly.

    Args:
        request: Generation request with description and constraints
        api_key: Validated API key with chat:write scope
        tenant_id: Tenant ID for RAG retrieval

    Returns:
        OptimizedGenerationResponse with generated code and cost metrics
    """
    try:
        from ..services.generation import (
            get_optimized_generation_service,
            ModelTier,
        )

        service = get_optimized_generation_service()

        # Convert request
        gen_request = FunctionGenerationRequest(
            description=request.description,
            runtime=request.runtime,
            inputs=request.inputs,
            outputs=request.outputs,
            constraints=request.constraints,
            examples=request.examples,
        )

        # Determine tier override
        force_tier = None
        if request.force_tier:
            try:
                force_tier = ModelTier(request.force_tier.lower())
            except ValueError:
                raise HTTPException(
                    status_code=status.HTTP_400_BAD_REQUEST,
                    detail=f"Invalid force_tier: {request.force_tier}. Use: cheap, mid, premium",
                )

        # Generate with optimization
        response, metrics = await service.generate(
            request=gen_request,
            tenant_id=tenant_id,
            api_key_info=api_key,
            force_tier=force_tier,
        )

        # Build optimization notes
        notes = []
        if metrics.cache_hit:
            notes.append("Served from cache - zero AI cost")
        elif metrics.template_used:
            notes.append("Template-based generation used - reduced token usage")
        if metrics.savings_vs_premium_pct > 50:
            notes.append(f"Saved {metrics.savings_vs_premium_pct:.0f}% vs premium model")
        if metrics.total_attempts > 1:
            notes.append(f"Auto-escalated through {metrics.total_attempts} tiers for quality")

        return OptimizedGenerationResponse(
            success=response.success,
            result=response.result,
            error=response.error,
            generation_id=response.generation_id,
            latency_ms=response.latency_ms,
            tokens_used=response.tokens_used,
            metrics=OptimizedGenerationMetricsResponse(
                total_attempts=metrics.total_attempts,
                final_tier=metrics.final_tier,
                cache_hit=metrics.cache_hit,
                template_used=metrics.template_used,
                validation_attempts=metrics.validation_attempts,
                total_cost_usd=metrics.total_cost_usd,
                savings_vs_premium_usd=metrics.savings_vs_premium_usd,
                savings_vs_premium_pct=metrics.savings_vs_premium_pct,
            ),
            optimization_notes=notes,
        )

    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Optimized generation failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=sanitize_error_message(e, include_details=settings.debug),
        )


@router.get("/api/composer/optimized-stats")
async def get_optimized_generation_stats(
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.CHAT_WRITE)),
):
    """Get statistics for the optimized generation service.

    Returns:
        Cache stats, cost tracking, and optimization metrics
    """
    try:
        from ..services.generation import get_optimized_generation_service

        service = get_optimized_generation_service()
        stats = await service.get_stats()

        return {
            "cache": stats["cache"],
            "costs": stats["costs"],
            "optimization_enabled": True,
            "strategies": [
                "multi_tier_routing",
                "template_rag_retrieval",
                "validation_pipeline",
                "intelligent_caching",
                "auto_escalation",
            ],
        }

    except Exception as e:
        logger.error(f"Get optimized stats failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to get optimization stats",
        )


# =============================================================================
# AI Composer aliases under /api/ai/composer namespace
# =============================================================================


@router.post("/api/ai/composer/generate", response_model=FunctionGenerationResponse)
async def ai_composer_generate_alias(
    request: FunctionGenerationRequest,
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.CHAT_WRITE)),
):
    """Alias for function generation under /api/ai/composer namespace."""
    return await generate_function(request, api_key)


@router.post("/api/ai/composer/generate/stream")
async def ai_composer_generate_stream_alias(
    request: FunctionGenerationRequest,
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.CHAT_WRITE)),
):
    """Streaming alias for function generation under /api/ai/composer namespace."""
    return await generate_function_stream(request, api_key)


@router.post("/api/ai/composer/generate-optimized", response_model=OptimizedGenerationResponse)
async def ai_composer_generate_optimized_alias(
    request: OptimizedGenerationRequest,
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.CHAT_WRITE)),
    tenant_id: str = Query("default", description="Tenant ID"),
):
    """Optimized generation alias under /api/ai/composer namespace."""
    return await generate_function_optimized(request, api_key, tenant_id)


@router.post("/api/ai/composer/refine")
async def ai_composer_refine(
    generation_id: str,
    modification_request: str = Query(..., description="What to change"),
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.CHAT_WRITE)),
):
    """Refine an existing AI-generated function - RESERVED FOR FUTURE."""
    return {
        "feature": "composer_refine",
        "status": "coming_soon",
        "message": "Refinement API is coming soon. Use chat interface for now.",
        "generation_id": generation_id,
        "modification_request": modification_request,
    }
