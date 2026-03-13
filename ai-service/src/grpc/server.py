"""gRPC server for FlyMind service.

Implements FlyMindService by delegating to routing, prewarming, caching,
moderation, and health services. Requires generated stubs; run from ai-service:

  python scripts/generate_grpc.py

Then start the server (e.g. via main app or standalone). Use TLS and
auth interceptors in production.
"""

import logging
import time
from typing import Optional

logger = logging.getLogger(__name__)

# Optional generated stubs (run scripts/generate_grpc.py to create)
try:
    from . import flymind_pb2
    from . import flymind_pb2_grpc
    from google.protobuf import timestamp_pb2
    from google.protobuf import empty_pb2
    _GRPC_AVAILABLE = True
except ImportError as e:
    flymind_pb2 = None
    flymind_pb2_grpc = None
    timestamp_pb2 = None
    empty_pb2 = None
    _GRPC_AVAILABLE = False
    _GRPC_IMPORT_ERROR = str(e)


if _GRPC_AVAILABLE:

    def _edge_to_target_platform(edge_value: str) -> int:
        """Map EdgeProvider value to proto TargetPlatform."""
        from ..models.schemas import EdgeProvider
        m = {
            EdgeProvider.CLOUDFLARE.value: flymind_pb2.TARGET_PLATFORM_CLOUDFLARE,
            EdgeProvider.VERCEL.value: flymind_pb2.TARGET_PLATFORM_VERCEL,
            EdgeProvider.FLY.value: flymind_pb2.TARGET_PLATFORM_FLY,
            EdgeProvider.DENO.value: flymind_pb2.TARGET_PLATFORM_DENO,
            EdgeProvider.FUNCTIONFLY.value: flymind_pb2.TARGET_PLATFORM_FUNCTIONFLY,
        }
        return m.get(edge_value, flymind_pb2.TARGET_PLATFORM_UNSPECIFIED)


if _GRPC_AVAILABLE:

    class FlyMindServicer(flymind_pb2_grpc.FlyMindServiceServicer):
        """Implements FlyMindService by delegating to existing services."""

        async def GetRoutingRecommendation(self, request, context):
            from ..services.routing import get_routing_service
            from ..models.schemas import RoutingDecisionRequest

            req = RoutingDecisionRequest(
                function_id=request.function_id,
                user_geography=request.geography or None,
                user_latitude=request.user_latitude if request.user_latitude else None,
                user_longitude=request.user_longitude if request.user_longitude else None,
                metadata=dict(request.attributes) if request.attributes else {},
            )
            routing = get_routing_service()
            decision = await routing.decide_routing(req)
            return flymind_pb2.RoutingRecommendation(
                target=_edge_to_target_platform(decision.recommended_edge.value),
                confidence=decision.confidence,
                reasoning=decision.reasoning,
                alternatives=[_edge_to_target_platform(a.value) for a in decision.alternatives],
                latency_estimate_ms=decision.latency_estimate_ms,
            )

        async def ReportExecutionOutcome(self, request, context):
            # Optionally persist or forward to analytics
            return empty_pb2.Empty()

        async def ShouldPrewarm(self, request, context):
            from ..services.prewarming import get_forecasting_service

            forecaster = get_forecasting_service()
            should_prewarm = await forecaster.should_prewarm(request.function_id)
            instances = 1 if should_prewarm else 0
            return flymind_pb2.PrewarmDecision(
                should_prewarm=should_prewarm,
                instances=instances,
                confidence=0.8 if should_prewarm else 0.2,
                reasoning="Forecast-based prewarm recommendation",
            )

        async def GetCacheKey(self, request, context):
            from ..config import settings

            # Simple policy: allow cache with default TTL
            return flymind_pb2.CacheDecision(
                should_cache=True,
                cache_key=request.input_hash.hex() if request.input_hash else "",
                ttl_seconds=settings.redis_cache_ttl,
            )

        async def ModerateContent(self, request, context):
            from ..services.moderation import get_moderation_service

            tenant_id = request.context.tenant_id if request.context else None
            content = request.content.decode("utf-8", errors="replace")
            content_type = "text"
            if request.content_type == flymind_pb2.CONTENT_TYPE_INPUT:
                content_type = "text"
            elif request.content_type == flymind_pb2.CONTENT_TYPE_OUTPUT:
                content_type = "text"

            mod = get_moderation_service()
            result = await mod.scan(
                content=content,
                content_type=content_type,
                tenant_id=tenant_id or None,
                user_id=request.context.principal_id if request.context else None,
            )
            violations = [
                flymind_pb2.ModerationViolation(
                    category=v.category.value,
                    message=v.message,
                    confidence=v.confidence,
                )
                for v in result.violations
            ]
            return flymind_pb2.ModerationResponse(
                allowed=result.is_allowed,
                violations=violations,
                risk_score=result.severity_score,
            )

        async def HealthCheck(self, request, context):
            from ..observability.health import get_health_checker
            from ..config import settings

            checker = get_health_checker()
            await checker.check_all()
            overall = checker.get_overall_status()
            healthy = overall.value == "healthy"
            now = timestamp_pb2.Timestamp()
            now.seconds = int(time.time())
            now.nanos = 0
            return flymind_pb2.HealthResponse(
                healthy=healthy,
                version=getattr(settings, "service_version", "1.0.0"),
                timestamp=now,
            )


class FlyMindGrpcServer:
    """gRPC server for FlyMind service."""

    def __init__(self, host: str = "0.0.0.0", port: int = 50051):
        self.host = host
        self.port = port
        self._server: Optional[object] = None

    async def start(self):
        """Start the gRPC server if stubs are available."""
        if not _GRPC_AVAILABLE:
            logger.warning(
                "gRPC stubs not found. Run: python scripts/generate_grpc.py (from ai-service). %s",
                _GRPC_IMPORT_ERROR,
            )
            return
        import grpc
        from grpc import aio

        self._server = aio.server()
        flymind_pb2_grpc.add_FlyMindServiceServicer_to_server(
            FlyMindServicer(), self._server
        )
        self._server.add_insecure_port(f"{self.host}:{self.port}")
        await self._server.start()
        logger.info("gRPC server listening on %s:%s", self.host, self.port)

    async def stop(self):
        """Stop the gRPC server."""
        if self._server is not None:
            await self._server.stop(0)
            self._server = None
            logger.info("gRPC server stopped")

    def is_running(self) -> bool:
        return self._server is not None


_grpc_server: Optional[FlyMindGrpcServer] = None


def get_grpc_server() -> FlyMindGrpcServer:
    global _grpc_server
    if _grpc_server is None:
        _grpc_server = FlyMindGrpcServer()
    return _grpc_server
