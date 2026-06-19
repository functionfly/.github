"""Debugging endpoints for error analysis and fix suggestions."""

import logging

from fastapi import APIRouter, HTTPException, status, Depends

from ..security.auth import (
    require_api_key_with_scope,
    APIKeyInfo,
    KeyScope,
)
from ..models.schemas import (
    DebugAnalyzeRequest,
    DebugAnalysis,
    DebugSuggestResponse,
    FixSuggestion,
)
from ..services.debugging import get_error_analyzer, get_fix_suggester

logger = logging.getLogger(__name__)

router = APIRouter()


@router.post("/api/debug/analyze", response_model=DebugAnalysis)
async def analyze_error(
    request: DebugAnalyzeRequest,
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.CHAT_READ)),
):
    """Analyze an error and provide root cause analysis.

    Args:
        request: Debug analysis request
        api_key: Validated API key with chat:read scope

    Returns:
        DebugAnalysis with root cause and suggestions
    """
    try:
        analyzer = get_error_analyzer()
        analysis = await analyzer.analyze_error(
            function_id=request.function_id,
            error_message=request.error_message,
            stack_trace=request.stack_trace,
        )
        return DebugAnalysis(**analysis)
    except Exception as e:
        logger.error(f"Debug analysis failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to analyze error",
        )


@router.post("/api/debug/suggest", response_model=DebugSuggestResponse)
async def get_fix_suggestions(
    request: DebugAnalyzeRequest,
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.CHAT_READ)),
):
    """Get fix suggestions for an error.

    Args:
        request: Debug analysis request
        api_key: Validated API key with chat:read scope

    Returns:
        DebugSuggestResponse with suggestions
    """
    try:
        analyzer = get_error_analyzer()
        analysis = await analyzer.analyze_error(
            function_id=request.function_id,
            error_message=request.error_message,
            stack_trace=request.stack_trace,
        )

        suggester = get_fix_suggester()
        suggestions = await suggester.generate_suggestions(analysis)
        docs = suggester.get_documentation_links(analysis.get("error_category", ""))

        return DebugSuggestResponse(
            analysis=DebugAnalysis(**analysis),
            suggestions=[FixSuggestion(**s) for s in suggestions],
            documentation_links=docs,
        )
    except Exception as e:
        logger.error(f"Get suggestions failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to get fix suggestions",
        )
