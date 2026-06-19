"""AI Graph Composition endpoints - Backend as a Graph."""

import logging
from typing import Optional

from fastapi import APIRouter, HTTPException, Query, status, Depends

from ..security.auth import (
    require_api_key_with_scope,
    APIKeyInfo,
    KeyScope,
)
from ..utils.security import sanitize_error_message
from ..config import settings
from ..models.schemas import (
    GraphCompositionRequest,
    GraphCompositionResponse,
    GraphTemplateListResponse,
)

logger = logging.getLogger(__name__)

router = APIRouter()


@router.get("/api/composition/templates", response_model=GraphTemplateListResponse)
async def list_graph_templates():
    """List available prebuilt graph templates.

    Returns all available templates for common backend patterns:
    - SaaS Starter (auth, billing, email)
    - E-commerce Checkout
    - API Backend (CRUD, auth, caching)
    - Webhook Processor

    Returns:
        GraphTemplateListResponse with all templates
    """
    try:
        from ..services.graph_composition import get_graph_composition_service

        service = get_graph_composition_service()
        templates = service.list_templates()

        return GraphTemplateListResponse(
            templates=templates,
            total_count=len(templates),
        )
    except Exception as e:
        logger.error(f"Failed to list templates: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=sanitize_error_message(e, include_details=settings.debug),
        )


@router.post("/api/composition/compose", response_model=GraphCompositionResponse)
async def compose_graph_from_prompt(
    request: GraphCompositionRequest,
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.CHAT_WRITE)),
):
    """Generate a graph using AI composition from natural language.

    This is the core "Backend as a Graph" feature. Users describe what they want
    (e.g., "Create a SaaS signup flow with Stripe billing and welcome email")
    and AI generates a complete graph definition with nodes, edges, and triggers.

    The service will:
    1. Match against templates if applicable
    2. Use LLM to generate topology for custom workflows
    3. Suggest function nodes from the catalog
    4. Connect them with appropriate data flows

    Example prompts:
    - "SaaS signup: validate email, create Stripe customer, send welcome email"
    - "E-commerce checkout: validate cart, process payment, create order, send receipt"
    - "API backend for blog with auth, CRUD, and caching"

    Args:
        request: Composition request with prompt and requirements
        api_key: Validated API key with chat:write scope

    Returns:
        GraphCompositionResponse with complete graph definition
    """
    try:
        from ..services.graph_composition import get_graph_composition_service

        service = get_graph_composition_service()
        response = await service.compose_from_prompt(
            request=request,
            api_key_info=api_key,
        )
        return response
    except Exception as e:
        logger.error(f"Graph composition failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=sanitize_error_message(e, include_details=settings.debug),
        )


@router.post("/api/composition/template/{template_id}", response_model=GraphCompositionResponse)
async def instantiate_graph_template(
    template_id: str,
    customization: Optional[str] = Query(None, description="Optional customization instructions"),
    tenant_id: Optional[str] = Query(None, description="Tenant ID"),
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.CHAT_WRITE)),
):
    """Instantiate a prebuilt graph template.

    Quick-start graph creation using prebuilt templates:
    - saas_starter: Auth + Stripe + Email
    - ecommerce_checkout: Cart validation -> Payment -> Order -> Receipt
    - api_backend: Auth -> Cache -> DB -> Response
    - webhook_processor: Signature validation -> Parsing -> Queue -> Processing

    Args:
        template_id: Template identifier (e.g., 'saas_starter')
        customization: Optional customization instructions
        tenant_id: Tenant ID for context
        api_key: Validated API key with chat:write scope

    Returns:
        GraphCompositionResponse with instantiated graph
    """
    try:
        from ..services.graph_composition import get_graph_composition_service

        service = get_graph_composition_service()

        # Check if template exists first
        if not service.get_template(template_id):
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=f"Template not found: {template_id}. Use /api/composition/templates to list available templates.",
            )

        response = await service.compose_from_template(
            template_id=template_id,
            customization_prompt=customization,
            tenant_id=tenant_id,
        )
        return response
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Template instantiation failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=sanitize_error_message(e, include_details=settings.debug),
        )
