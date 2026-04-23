"""Optimization endpoints for function performance recommendations."""

import logging

from fastapi import APIRouter, HTTPException, status

from ..models.schemas import (
    Recommendation,
    OptimizationRecommendationsResponse,
    ApplyRecommendationRequest,
    ApplyRecommendationResponse,
)
from ..services.optimization import get_recommendation_engine

logger = logging.getLogger(__name__)

router = APIRouter()


@router.get("/api/optimize/{function_id}", response_model=OptimizationRecommendationsResponse)
async def get_optimization_recommendations(function_id: str):
    """Get optimization recommendations for a function.

    Args:
        function_id: The function ID

    Returns:
        OptimizationRecommendationsResponse with recommendations
    """
    try:
        recommender = get_recommendation_engine()
        recommendations = await recommender.generate_recommendations(function_id)

        # Convert recommendations to proper format
        formatted_recommendations = []
        for i, rec in enumerate(recommendations):
            formatted_recommendations.append(
                Recommendation(
                    id=f"rec-{i}",
                    type=rec.get("type", ""),
                    title=rec.get("title", ""),
                    description=rec.get("description", ""),
                    category=rec.get("category", ""),
                    priority=rec.get("priority", "medium"),
                    action=rec.get("action", ""),
                    current_value=rec.get("current_value", 0),
                    target_value=rec.get("target_value", 0),
                    estimated_savings_monthly=rec.get("estimated_savings_monthly", 0),
                )
            )

        return OptimizationRecommendationsResponse(
            function_id=function_id,
            recommendations=formatted_recommendations,
            total_count=len(formatted_recommendations),
        )
    except Exception as e:
        logger.error(f"Get recommendations failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to get recommendations",
        )


@router.post("/api/optimize/{function_id}/apply", response_model=ApplyRecommendationResponse)
async def apply_optimization_recommendation(
    function_id: str,
    request: ApplyRecommendationRequest,
):
    """Apply an optimization recommendation.

    Args:
        function_id: The function ID
        request: Apply recommendation request

    Returns:
        ApplyRecommendationResponse with result
    """
    try:
        recommender = get_recommendation_engine()
        result = await recommender.apply_recommendation(
            function_id=function_id,
            recommendation_id=request.recommendation_id,
        )
        return ApplyRecommendationResponse(**result)
    except Exception as e:
        logger.error(f"Apply recommendation failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to apply recommendation",
        )
