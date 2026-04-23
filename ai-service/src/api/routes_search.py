"""Search endpoints for semantic function search."""

import logging

from fastapi import APIRouter, HTTPException, Query, status

from ..models.schemas import (
    SearchQuery,
    SearchResult,
    SearchResponse,
)
from ..services.search import get_search_indexer, get_result_ranker, get_query_processor
from ..services.flyembed import get_flyembed_service

logger = logging.getLogger(__name__)

router = APIRouter()


@router.post("/api/search/functions", response_model=SearchResponse)
async def search_functions(
    request: SearchQuery,
    tenant_id: str = Query(..., description="Tenant ID"),
):
    """Search functions using semantic search.

    Args:
        request: Search query
        tenant_id: The tenant ID

    Returns:
        SearchResponse with results
    """
    try:
        query_processor = get_query_processor()
        indexer = get_search_indexer()
        ranker = get_result_ranker()

        if request.use_triple:
            # Triple-vector search path
            processed_query, _, metadata = await query_processor.process_query(request.query)

            # Generate triple query vectors
            flyembed = get_flyembed_service()
            triple_query = await flyembed.embed_query(request.query)

            results = await indexer.search_triple(
                tenant_id=tenant_id,
                query_vectors={
                    "contract_vector": triple_query.contract_vector,
                    "semantic_vector": triple_query.semantic_vector,
                    "code_vector": triple_query.code_vector,
                },
                limit=request.limit,
            )

            if request.filters:
                results = ranker.filter_results(results, request.filters)
            ranked_results = ranker.rank_results_triple(results, request.query, request.weights)
        else:
            # Single-vector search path (legacy)
            processed_query, embedding, metadata = await query_processor.process_query(
                request.query
            )
            results = await indexer.search(
                tenant_id=tenant_id,
                query_embedding=embedding,
                limit=request.limit,
            )
            if request.filters:
                results = ranker.filter_results(results, request.filters)
            ranked_results = ranker.rank_results(results, request.query)

        # Keyword fallback if no results
        if not ranked_results:
            all_functions = await indexer.index_tenant_functions(tenant_id)
            ranked_results = query_processor.process_keyword_fallback(
                request.query,
                [{"data": f} for f in all_functions],
            )

        return SearchResponse(
            query=request.query,
            results=[
                SearchResult(
                    function_id=r["function_id"],
                    function_name=r["data"].get("name", ""),
                    description=r["data"].get("description"),
                    runtime=r["data"].get("runtime"),
                    tags=r["data"].get("tags", []),
                    score=r.get("score", 0),
                    rank=r.get("rank", 0),
                )
                for r in ranked_results[: request.limit]
            ],
            total_count=len(ranked_results),
            query_type=metadata.get("search_type", "semantic"),
        )
    except Exception as e:
        logger.error(f"Search failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to search functions",
        )
