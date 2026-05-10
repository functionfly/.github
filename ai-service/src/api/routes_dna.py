"""DNA Analysis Routes — AI-powered function evolution.

Provides endpoints for analyzing function execution patterns and generating
optimized code variants. Called by the Go DNA service worker.
"""

import difflib
import hashlib
import logging
from typing import Optional

from fastapi import APIRouter, HTTPException, Depends
from pydantic import BaseModel, field_validator

from ..providers.manager import get_provider_manager
from ..providers.base import Message, Role
from ..security.auth import require_api_key, APIKeyInfo

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/api/dna", tags=["dna"])

# Maximum allowed size for source code input (50KB)
MAX_CODE_SIZE = 50 * 1024


# ──────────────────────────────────────────────────────────────────────────────
# Request / Response Models
# ──────────────────────────────────────────────────────────────────────────────


class InputPattern(BaseModel):
    shape: str
    hash: str
    frequency: float
    count: int


class AggregatedMetrics(BaseModel):
    total_executions: int
    avg_latency_ms: float
    p50_latency_ms: float
    p95_latency_ms: float
    p99_latency_ms: float
    success_rate: float
    error_distribution: dict[str, int]
    cold_start_rate: float
    avg_memory_peak_mb: float
    input_patterns: Optional[list[InputPattern]] = None


class AnalyzeRequest(BaseModel):
    function_id: str
    mutation_type: str
    trigger_reason: str
    metrics: AggregatedMetrics
    current_code: Optional[str] = None

    @field_validator('current_code')
    @classmethod
    def validate_code_size(cls, v):
        if v is not None and len(v) > MAX_CODE_SIZE:
            raise ValueError(
                f"Source code size ({len(v)} bytes) exceeds maximum allowed size ({MAX_CODE_SIZE} bytes)."
            )
        return v


class AnalyzeResponse(BaseModel):
    should_mutate: bool
    mutation_type: str
    severity: str  # low, medium, high, critical
    analysis: str
    recommended_action: str


class GenerateRequest(BaseModel):
    function_id: str
    mutation_type: str
    trigger_reason: str
    metrics: AggregatedMetrics
    current_code: Optional[str] = None
    runtime: Optional[str] = None

    @field_validator('current_code')
    @classmethod
    def validate_code_size(cls, v):
        if v is not None and len(v) > MAX_CODE_SIZE:
            raise ValueError(
                f"Source code size ({len(v)} bytes) exceeds maximum allowed size ({MAX_CODE_SIZE} bytes). "
                f"Please reduce the function size or split into smaller functions."
            )
        return v


class GenerateResponse(BaseModel):
    original_code: Optional[str] = None
    mutated_code: Optional[str] = None
    original_hash: Optional[str] = None
    mutated_hash: Optional[str] = None
    diff: Optional[str] = None
    estimated_impact: dict = {}
    confidence: float = 0.0
    model_used: str = ""


# ──────────────────────────────────────────────────────────────────────────────
# Endpoints
# ──────────────────────────────────────────────────────────────────────────────


MUTATION_PROMPTS = {
    "optimize_latency": (
        "Optimize this function for lower latency. Focus on:\n"
        "- Reducing unnecessary I/O or computation\n"
        "- Caching repeated lookups\n"
        "- Using more efficient algorithms or data structures\n"
        "- Minimizing memory allocations in hot paths\n"
        "Do NOT change the function signature or external behavior."
    ),
    "reduce_memory": (
        "Optimize this function for lower memory usage. Focus on:\n"
        "- Reducing peak memory allocations\n"
        "- Streaming data instead of buffering\n"
        "- Freeing resources earlier\n"
        "- Using memory-efficient data structures\n"
        "Do NOT change the function signature or external behavior."
    ),
    "fix_error_pattern": (
        "Improve this function's reliability and error handling. Focus on:\n"
        "- Adding robust error handling for known failure modes\n"
        "- Adding retry logic for transient failures\n"
        "- Validating inputs more carefully\n"
        "- Adding graceful degradation paths\n"
        "Do NOT change the function signature or external behavior."
    ),
    "improve_reliability": (
        "Improve this function's reliability. Focus on:\n"
        "- Making the function more resilient to edge cases\n"
        "- Adding defensive checks and fallback paths\n"
        "- Improving timeout handling\n"
        "- Reducing the blast radius of failures\n"
        "Do NOT change the function signature or external behavior."
    ),
    "refactor_hotpath": (
        "Refactor this function for better overall performance. Focus on:\n"
        "- Identifying and optimizing hot paths\n"
        "- Reducing cyclomatic complexity\n"
        "- Eliminating redundant code paths\n"
        "- Improving algorithmic complexity\n"
        "Do NOT change the function signature or external behavior."
    ),
}


def _compute_diff(original: str, mutated: str) -> str:
    """Compute a unified diff between original and mutated code."""
    original_lines = original.splitlines(keepends=True)
    mutated_lines = mutated.splitlines(keepends=True)
    diff = difflib.unified_diff(
        original_lines,
        mutated_lines,
        fromfile="original",
        tofile="mutated",
        lineterm="",
    )
    return "".join(diff)


@router.post("/analyze", response_model=AnalyzeResponse)
async def analyze_function(
    req: AnalyzeRequest,
    _key: APIKeyInfo = Depends(require_api_key),
):
    """Analyze execution metrics and determine if a mutation is warranted.

    Returns a structured analysis with severity, recommended action, and
    whether the function should be mutated.
    """
    logger.info(f"DNA analysis requested for {req.function_id} (type={req.mutation_type})")

    severity = "low"
    analysis_parts: list[str] = []
    recommended = "no action needed"

    # Latency analysis
    if req.metrics.p99_latency_ms > 1000:
        severity = "high"
        analysis_parts.append(
            f"P99 latency is {req.metrics.p99_latency_ms:.0f}ms (critical threshold: 1000ms). "
            f"Hot path optimization recommended."
        )
        recommended = "optimize_latency"
    elif req.metrics.p99_latency_ms > 500:
        severity = max(severity, "medium")
        analysis_parts.append(
            f"P99 latency is {req.metrics.p99_latency_ms:.0f}ms (warning threshold: 500ms)."
        )
        recommended = "optimize_latency"

    # Error rate analysis
    error_rate = 1.0 - req.metrics.success_rate
    if error_rate > 0.05:
        severity = "high"
        top_errors = sorted(
            req.metrics.error_distribution.items(), key=lambda x: x[1], reverse=True
        )[:3]
        error_summary = ", ".join(f"{k}: {v}" for k, v in top_errors)
        analysis_parts.append(
            f"Error rate is {error_rate * 100:.1f}% (threshold: 5%). "
            f"Top errors: {error_summary}"
        )
        recommended = "fix_error_pattern"

    # Cold start analysis
    if req.metrics.cold_start_rate > 0.3:
        severity = max(severity, "medium")
        analysis_parts.append(
            f"Cold start rate is {req.metrics.cold_start_rate * 100:.1f}% (threshold: 30%). "
            f"Memory optimization may reduce cold starts."
        )
        if recommended == "no action needed":
            recommended = "reduce_memory"

    # Memory analysis
    if req.metrics.avg_memory_peak_mb > 256:
        severity = max(severity, "medium")
        analysis_parts.append(
            f"Average peak memory is {req.metrics.avg_memory_peak_mb:.0f}MB (threshold: 256MB)."
        )
        if recommended == "no action needed":
            recommended = "reduce_memory"

    should_mutate = len(analysis_parts) > 0
    analysis_text = " | ".join(analysis_parts) if analysis_parts else "Function is performing within normal parameters."

    return AnalyzeResponse(
        should_mutate=should_mutate,
        mutation_type=req.mutation_type if should_mutate else "none",
        severity=severity,
        analysis=analysis_text,
        recommended_action=recommended,
    )


@router.post("/generate", response_model=GenerateResponse)
async def generate_variant(
    req: GenerateRequest,
    _key: APIKeyInfo = Depends(require_api_key),
):
    """Generate an optimized code variant based on execution metrics.

    When source code is provided, uses the LLM to generate an optimized variant.
    Without source code, returns a structured proposal with estimated impact only.
    """
    logger.info(
        f"DNA variant generation requested for {req.function_id} "
        f"(type={req.mutation_type}, executions={req.metrics.total_executions}, "
        f"has_code={req.current_code is not None})"
    )

    # Compute confidence based on data volume and signal strength
    confidence = 0.5
    if req.metrics.total_executions > 10000:
        confidence += 0.15
    if req.metrics.total_executions > 50000:
        confidence += 0.1
    if req.metrics.success_rate < 0.95:
        confidence += 0.1  # Strong error signal
    if req.metrics.p99_latency_ms > 1000:
        confidence += 0.1  # Strong latency signal
    confidence = min(confidence, 0.95)

    # Estimate impact based on mutation type
    estimated_impact = {
        "latency_improvement_pct": 0.0,
        "memory_reduction_pct": 0.0,
        "reliability_improvement_pct": 0.0,
    }

    if req.mutation_type == "optimize_latency":
        excess = max(0, req.metrics.p99_latency_ms - 200) / req.metrics.p99_latency_ms
        estimated_impact["latency_improvement_pct"] = round(min(excess * 50, 60), 1)
    elif req.mutation_type == "reduce_memory":
        excess = max(0, req.metrics.avg_memory_peak_mb - 128) / max(req.metrics.avg_memory_peak_mb, 1)
        estimated_impact["memory_reduction_pct"] = round(min(excess * 40, 50), 1)
    elif req.mutation_type == "fix_error_pattern":
        error_rate = (1.0 - req.metrics.success_rate) * 100
        estimated_impact["reliability_improvement_pct"] = round(min(error_rate * 2, 15), 1)
    elif req.mutation_type == "improve_reliability":
        estimated_impact["reliability_improvement_pct"] = round(
            min((1.0 - req.metrics.success_rate) * 100 * 1.5, 10), 1
        )
    elif req.mutation_type == "refactor_hotpath":
        estimated_impact["latency_improvement_pct"] = 15.0
        estimated_impact["memory_reduction_pct"] = 10.0

    # If source code is provided, call the LLM for real code generation
    if req.current_code:
        try:
            result = await _generate_with_llm(req)
            if result:
                return result
        except Exception as e:
            logger.warning(f"LLM code generation failed for {req.function_id}: {e}")
            # Fall through to return proposal-only response

    # No source code available — return proposal-only (no actual code)
    hash_input = f"{req.function_id}:{req.mutation_type}:{req.metrics.total_executions}"
    proposal_hash = hashlib.sha256(hash_input.encode()).hexdigest()[:16]

    return GenerateResponse(
        original_code=None,
        mutated_code=None,
        original_hash=proposal_hash,
        mutated_hash=None,
        diff=None,
        estimated_impact=estimated_impact,
        confidence=round(confidence, 2),
        model_used="flymind-analysis-v1",
    )


async def _generate_with_llm(req: GenerateRequest) -> Optional[GenerateResponse]:
    """Call the LLM to generate an optimized code variant."""
    manager = get_provider_manager()
    provider = manager.get_provider_for_chat("openrouter")
    if not provider:
        provider = manager.get_provider_for_chat("openai")
    if not provider:
        logger.warning("No LLM provider available for DNA code generation")
        return None

    mutation_prompt = MUTATION_PROMPTS.get(
        req.mutation_type,
        "Optimize this function for better performance and reliability."
    )

    # Build metrics summary for the LLM context
    metrics_summary = (
        f"Execution Metrics ({req.metrics.total_executions:,} executions):\n"
        f"- P99 Latency: {req.metrics.p99_latency_ms:.0f}ms\n"
        f"- Avg Latency: {req.metrics.avg_latency_ms:.0f}ms\n"
        f"- Success Rate: {req.metrics.success_rate * 100:.1f}%\n"
        f"- Cold Start Rate: {req.metrics.cold_start_rate * 100:.1f}%\n"
        f"- Avg Peak Memory: {req.metrics.avg_memory_peak_mb:.0f}MB\n"
    )
    if req.metrics.error_distribution:
        top_errors = sorted(req.metrics.error_distribution.items(), key=lambda x: x[1], reverse=True)[:5]
        metrics_summary += "- Top Errors: " + ", ".join(f"{k}({v})" for k, v in top_errors) + "\n"

    system_prompt = (
        "You are an expert code optimizer. You will receive function source code and "
        "production execution metrics. Generate an optimized version of the function.\n\n"
        "Rules:\n"
        "- Preserve the function signature exactly\n"
        "- Preserve all external behavior and return types\n"
        "- Only optimize for the specific metric that needs improvement\n"
        "- Return ONLY the optimized function code, no explanations\n"
        "- Do NOT add comments explaining changes\n"
        "- Ensure the code is syntactically valid\n"
    )

    user_prompt = (
        f"{mutation_prompt}\n\n"
        f"Function source code:\n```\n{req.current_code}\n```\n\n"
        f"{metrics_summary}\n"
        f"Trigger reason: {req.trigger_reason}\n\n"
        f"Return ONLY the optimized function code."
    )

    messages = [
        Message(role=Role.SYSTEM, content=system_prompt),
        Message(role=Role.USER, content=user_prompt),
    ]

    response = await provider.complete(
        messages=messages,
        temperature=0.2,
        max_tokens=4000,
    )

    mutated_code = response.content.strip()
    if not mutated_code:
        return None

    # Strip markdown code fences if present
    if mutated_code.startswith("```"):
        lines = mutated_code.split("\n")
        # Remove first line (```python or similar) and last line (```)
        if lines[-1].strip() == "```":
            lines = lines[1:-1]
        elif lines[0].strip().startswith("```"):
            lines = lines[1:]
        mutated_code = "\n".join(lines)

    original_code = req.current_code
    original_hash = hashlib.sha256(original_code.encode()).hexdigest()[:16]
    mutated_hash = hashlib.sha256(mutated_code.encode()).hexdigest()[:16]
    diff = _compute_diff(original_code, mutated_code)

    model_name = "unknown"
    if hasattr(provider, "model"):
        model_name = provider.model
    provider_info = provider.get_provider_info()
    model_used = f"{provider_info.get('provider', 'unknown')}/{model_name}"

    return GenerateResponse(
        original_code=original_code,
        mutated_code=mutated_code,
        original_hash=original_hash,
        mutated_hash=mutated_hash,
        diff=diff,
        estimated_impact={},  # LLM-generated code doesn't need impact estimates
        confidence=round(min(0.85, 0.6 + (req.metrics.total_executions / 100000) * 0.25), 2),
        model_used=model_used,
    )
