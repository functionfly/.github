"""FlyEmbed triple-vector embedding endpoints."""

import logging

from fastapi import APIRouter, HTTPException, status, Depends

from ..security.auth import (
    require_api_key_with_scope,
    APIKeyInfo,
    KeyScope,
)
from ..models.schemas import (
    TripleEmbeddingRequest,
    TripleEmbeddingResult,
    TripleEmbeddingBatchRequest,
    TripleEmbeddingBatchResponse,
    TripleQueryRequest,
    TripleQueryVector,
)
from ..services.flyembed import get_flyembed_service

logger = logging.getLogger(__name__)

router = APIRouter()


@router.post("/api/flyembed/embed", response_model=TripleEmbeddingResult)
async def flyembed_embed(
    request: TripleEmbeddingRequest,
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.EMBED_WRITE)),
):
    """Generate triple embeddings for a single function.

    Args:
        request: Triple embedding request with function data
        api_key: Validated API key with embed:write scope

    Returns:
        TripleEmbeddingResult with contract, semantic, and code embeddings
    """
    try:
        service = get_flyembed_service()
        result = await service.embed_function(request.model_dump())
        return TripleEmbeddingResult(
            function_id=result.function_id,
            contract_embedding=result.contract_embedding,
            semantic_embedding=result.semantic_embedding,
            code_embedding=result.code_embedding,
            contract_text=result.contract_text,
            semantic_text=result.semantic_text,
            code_text=result.code_text,
            latency_ms=result.latency_ms,
        )
    except Exception as e:
        logger.error(f"FlyEmbed embed failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to generate triple embeddings: {e}",
        )


@router.post("/api/flyembed/embed-batch", response_model=TripleEmbeddingBatchResponse)
async def flyembed_embed_batch(
    request: TripleEmbeddingBatchRequest,
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.EMBED_ADMIN)),
):
    """Batch generate triple embeddings for multiple functions.

    Args:
        request: Batch of function data
        api_key: Validated API key with embed:admin scope (batch operations require elevated permissions)

    Returns:
        TripleEmbeddingBatchResponse with all results
    """
    try:
        service = get_flyembed_service()
        functions_data = [f.model_dump() for f in request.functions]
        results = await service.embed_batch(functions_data)
        return TripleEmbeddingBatchResponse(
            results=[
                TripleEmbeddingResult(
                    function_id=r.function_id,
                    contract_embedding=r.contract_embedding,
                    semantic_embedding=r.semantic_embedding,
                    code_embedding=r.code_embedding,
                    contract_text=r.contract_text,
                    semantic_text=r.semantic_text,
                    code_text=r.code_text,
                    latency_ms=r.latency_ms,
                )
                for r in results
            ],
            total_count=len(results),
        )
    except Exception as e:
        logger.error(f"FlyEmbed batch embed failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to batch generate triple embeddings: {e}",
        )


@router.post("/api/flyembed/query", response_model=TripleQueryVector)
async def flyembed_query(
    request: TripleQueryRequest,
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.EMBED_READ)),
):
    """Generate triple query vectors for search.

    Args:
        request: Query request with search text
        api_key: Validated API key with embed:read scope

    Returns:
        TripleQueryVector with three query vectors
    """
    try:
        service = get_flyembed_service()
        result = await service.embed_query(request.query)
        return TripleQueryVector(
            query=result.query,
            contract_vector=result.contract_vector,
            semantic_vector=result.semantic_vector,
            code_vector=result.code_vector,
            dimensions=512,
            latency_ms=result.latency_ms,
        )
    except Exception as e:
        logger.error(f"FlyEmbed query failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to generate triple query vectors: {e}",
        )


@router.get("/api/flyembed/health")
async def flyembed_health():
    """Health check for FlyEmbed service.

    Returns:
        Status of the FlyEmbed service
    """
    return {"status": "healthy", "service": "flyembed"}
