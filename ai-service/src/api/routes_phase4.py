


# =============================================================================
# Phase 4: Production Hardening - Moderation Endpoints
# =============================================================================

from pydantic import BaseModel
from typing import List, Dict, Any, Optional


class ModerationScanRequest(BaseModel):
    """Request to scan content for policy violations."""
    content: str
    content_type: str = "text"
    tenant_id: Optional[str] = None


class ModerationScanResponse(BaseModel):
    """Response from content moderation scan."""
    id: str
    decision: str
    violations: List[Dict[str, Any]]
    scanned_at: str
    scan_duration_ms: float


class PolicyRuleRequest(BaseModel):
    """Request to create/update a policy rule."""
    category: str
    action: str
    severity_threshold: str = "low"
    enabled: bool = True


class PolicyRequest(BaseModel):
    """Request to create/update a policy."""
    name: str
    description: Optional[str] = None
    rules: List[PolicyRuleRequest] = []


class PolicyResponse(BaseModel):
    """Policy response."""
    id: str
    name: str
    description: Optional[str]
    rules: List[Dict[str, Any]]
    is_default: bool
    is_active: bool
    created_at: str


@router.post("/api/moderate/scan", response_model=ModerationScanResponse)
async def scan_content(request: ModerationScanRequest):
    """Scan content for policy violations.

    Args:
        request: Moderation scan request

    Returns:
        ModerationScanResponse with scan results
    """
    try:
        from ..services.moderation import get_moderation_service

        service = get_moderation_service()
        result = await service.scan(
            content=request.content,
            content_type=request.content_type,
            tenant_id=request.tenant_id,
        )

        return ModerationScanResponse(
            id=result.id,
            decision=result.decision.value,
            violations=[
                {
                    "id": v.id,
                    "category": v.category.value,
                    "severity": v.severity,
                    "message": v.message,
                }
                for v in result.violations
            ],
            scanned_at=result.scanned_at.isoformat(),
            scan_duration_ms=result.scan_duration_ms,
        )
    except Exception as e:
        logger.error(f"Moderation scan failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to scan content",
        )


@router.get("/api/moderate/policies", response_model=List[PolicyResponse])
async def list_policies(tenant_id: Optional[str] = None):
    """List moderation policies.

    Args:
        tenant_id: Optional tenant ID to filter by

    Returns:
        List of policies
    """
    try:
        from ..services.moderation import get_policy_manager

        manager = get_policy_manager()
        policies = manager.list_policies(tenant_id=tenant_id)

        return [
            PolicyResponse(
                id=p.id,
                name=p.name,
                description=p.description,
                rules=[
                    {
                        "id": r.id,
                        "category": r.category.value,
                        "action": r.action.value,
                        "severity_threshold": r.severity_threshold,
                        "enabled": r.enabled,
                    }
                    for r in p.rules
                ],
                is_default=p.is_default,
                is_active=p.is_active,
                created_at=p.created_at.isoformat(),
            )
            for p in policies
        ]
    except Exception as e:
        logger.error(f"List policies failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to list policies",
        )


@router.post("/api/moderate/policies", response_model=PolicyResponse)
async def create_policy(request: PolicyRequest):
    """Create a moderation policy.

    Args:
        request: Policy creation request

    Returns:
        Created policy
    """
    try:
        from ..services.moderation import (
            get_policy_manager,
            PolicyRule,
            PolicyAction,
            ModerationCategory,
        )

        manager = get_policy_manager()

        # Convert rules
        rules = []
        for r in request.rules:
            try:
                category = ModerationCategory(r.category)
                action = PolicyAction(r.action)
                rules.append(PolicyRule(
                    category=category,
                    action=action,
                    severity_threshold=r.severity_threshold,
                    enabled=r.enabled,
                ))
            except ValueError:
                pass  # Skip invalid rules

        policy = manager.create_policy(
            name=request.name,
            description=request.description,
            rules=rules,
        )

        return PolicyResponse(
            id=policy.id,
            name=policy.name,
            description=policy.description,
            rules=[
                {
                    "id": r.id,
                    "category": r.category.value,
                    "action": r.action.value,
                    "severity_threshold": r.severity_threshold,
                    "enabled": r.enabled,
                }
                for r in policy.rules
            ],
            is_default=policy.is_default,
            is_active=policy.is_active,
            created_at=policy.created_at.isoformat(),
        )
    except Exception as e:
        logger.error(f"Create policy failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to create policy",
        )


# =============================================================================
# Phase 4: Production Hardening - Cache Endpoints
# =============================================================================


class CacheStatsResponse(BaseModel):
    """Cache statistics response."""
    hits: int
    misses: int
    hit_rate: float
    entry_count: int


class CacheInvalidateRequest(BaseModel):
    """Request to invalidate cache entries."""
    key: Optional[str] = None
    pattern: Optional[str] = None
    tags: Optional[List[str]] = None


@router.get("/api/cache/stats", response_model=CacheStatsResponse)
async def get_cache_stats():
    """Get cache statistics.

    Returns:
        Cache statistics
    """
    try:
        from ..services.caching import get_cache_service

        service = get_cache_service()
        stats = service.get_stats()

        return CacheStatsResponse(
            hits=stats.hits,
            misses=stats.misses,
            hit_rate=stats.hit_rate,
            entry_count=stats.entry_count,
        )
    except Exception as e:
        logger.error(f"Get cache stats failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to get cache stats",
        )


@router.post("/api/cache/invalidate")
async def invalidate_cache(request: CacheInvalidateRequest):
    """Invalidate cache entries.

    Args:
        request: Invalidation request

    Returns:
        Number of entries invalidated
    """
    try:
        from ..services.caching import get_cache_service

        service = get_cache_service()

        if request.key:
            invalidated = service.invalidate_by_key(request.key)
        elif request.pattern:
            invalidated = service.invalidate_by_pattern(request.pattern)
        elif request.tags:
            invalidated = service.invalidate_by_tags(request.tags)
        else:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Must specify key, pattern, or tags",
            )

        return {"invalidated": len(invalidated)}
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Cache invalidation failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to invalidate cache",
        )


# =============================================================================
# Phase 4: Production Hardening - Metrics Endpoints
# =============================================================================


@router.get("/api/metrics")
async def get_metrics():
    """Get Prometheus metrics.

    Returns:
        Prometheus metrics in text format
    """
    try:
        from ..observability.metrics import get_metrics_collector

        collector = get_metrics_collector()
        return {"metrics": collector.get_metrics_text()}
    except Exception as e:
        logger.error(f"Get metrics failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to get metrics",
        )


# =============================================================================
# Phase 4: Production Hardening - Health Endpoints
# =============================================================================


@router.get("/api/health/ready")
async def readiness_check():
    """Readiness check endpoint.

    Returns:
        Readiness status
    """
    try:
        from ..observability.health import get_health_checker

        checker = get_health_checker()
        health = await checker.check_all()

        overall = checker.get_overall_status()

        return {
            "ready": overall.value == "healthy",
            "status": overall.value,
            "components": {
                name: comp.status.value
                for name, comp in health.items()
            },
        }
    except Exception as e:
        logger.error(f"Readiness check failed: {e}")
        return {
            "ready": False,
            "status": "unhealthy",
            "error": str(e),
        }


@router.get("/api/health/live")
async def liveness_check():
    """Liveness check endpoint.

    Returns:
        Liveness status
    """
    return {"alive": True}


# =============================================================================
# Phase 4: Production Hardening - Rate Limiting Endpoints
# =============================================================================


@router.get("/api/rate-limit/usage")
async def get_rate_limit_usage(tenant_id: str):
    """Get rate limit usage for a tenant.

    Args:
        tenant_id: Tenant ID

    Returns:
        Rate limit usage
    """
    try:
        from ..middleware.rate_limiter import get_rate_limiter

        limiter = get_rate_limiter()
        usage = limiter.get_usage(tenant_id)

        return usage
    except Exception as e:
        logger.error(f"Get rate limit usage failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to get rate limit usage",
        )


# =============================================================================
# Phase 4: Production Hardening - Cost Tracking Endpoints
# =============================================================================


@router.get("/api/cost/usage")
async def get_cost_usage(tenant_id: str):
    """Get cost usage for a tenant.

    Args:
        tenant_id: Tenant ID

    Returns:
        Cost usage
    """
    try:
        from ..middleware.cost_tracker import get_cost_tracker

        tracker = get_cost_tracker()
        usage = tracker.get_usage(tenant_id)

        return usage
    except Exception as e:
        logger.error(f"Get cost usage failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to get cost usage",
        )


@router.get("/api/cost/by-provider")
async def get_cost_by_provider(tenant_id: str):
    """Get cost breakdown by provider.

    Args:
        tenant_id: Tenant ID

    Returns:
        Cost breakdown
    """
    try:
        from ..middleware.cost_tracker import get_cost_tracker

        tracker = get_cost_tracker()
        costs = tracker.get_costs_by_provider(tenant_id)

        return {"costs": costs}
    except Exception as e:
        logger.error(f"Get cost by provider failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to get cost breakdown",
        )


# =============================================================================
# Phase 4: Production Hardening - Background Tasks Endpoints
# =============================================================================


@router.get("/api/tasks")
async def list_tasks():
    """List background tasks.

    Returns:
        List of tasks
    """
    try:
        from ..workers.scheduler import get_task_scheduler

        scheduler = get_task_scheduler()
        tasks = scheduler.list_tasks()

        return {
            "tasks": [
                {
                    "id": t.id,
                    "name": t.name,
                    "enabled": t.enabled,
                    "interval_seconds": t.interval_seconds,
                    "last_run": t.last_run.isoformat() if t.last_run else None,
                    "run_count": t.run_count,
                    "error_count": t.error_count,
                }
                for t in tasks
            ]
        }
    except Exception as e:
        logger.error(f"List tasks failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to list tasks",
        )


@router.get("/api/tasks/stats")
async def get_task_stats():
    """Get task scheduler statistics.

    Returns:
        Task statistics
    """
    try:
        from ..workers.scheduler import get_task_scheduler

        scheduler = get_task_scheduler()
        stats = scheduler.get_stats()

        return stats
    except Exception as e:
        logger.error(f"Get task stats failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to get task stats",
        )


@router.post("/api/tasks/{task_id}/enable")
async def enable_task(task_id: str):
    """Enable a task.

    Args:
        task_id: Task ID

    Returns:
        Success status
    """
    try:
        from ..workers.scheduler import get_task_scheduler

        scheduler = get_task_scheduler()
        success = scheduler.enable_task(task_id)

        if not success:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="Task not found",
            )

        return {"success": True}
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Enable task failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to enable task",
        )


@router.post("/api/tasks/{task_id}/disable")
async def disable_task(task_id: str):
    """Disable a task.

    Args:
        task_id: Task ID

    Returns:
        Success status
    """
    try:
        from ..workers.scheduler import get_task_scheduler

        scheduler = get_task_scheduler()
        success = scheduler.disable_task(task_id)

        if not success:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="Task not found",
            )

        return {"success": True}
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Disable task failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to disable task",
        )

