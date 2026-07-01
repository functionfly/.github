"""Chat endpoints for the developer experience layer."""

import asyncio
import logging
from typing import Optional, List

from fastapi import APIRouter, HTTPException, Query, status, Body, Depends, Header
from pydantic import BaseModel

from ..security.auth import (
    require_api_key_with_scope,
    APIKeyInfo,
    KeyScope,
)
from ..models.schemas import (
    ChatSessionResponse,
    ChatMessageResponse,
    ChatHistoryResponse,
    ChatIntent,
    ChatMessage as SchemaChatMessage,
    ThinkingConfig,
)
from ..services.chat import get_chat_manager

logger = logging.getLogger(__name__)

router = APIRouter()


class ChatSendBody(BaseModel):
    session_id: str
    user_id: str
    message: str
    tenant_id: Optional[str] = None
    context: Optional[dict] = None


@router.post("/api/chat/message", response_model=ChatMessageResponse)
async def send_chat_message(
    session_id: Optional[str] = Query(
        None, description="Existing session ID (or empty to create new)"
    ),
    user_id: Optional[str] = Query(None, description="The user ID"),
    message: Optional[str] = Query(None, description="The message text"),
    tenant_id: Optional[str] = Query(
        None, description="Tenant ID for context (e.g. deployed functions)"
    ),
    payload: Optional[ChatSendBody] = Body(default=None),
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.CHAT_WRITE)),
    x_byok_key: Optional[str] = Header(default=None, alias="X-BYOK-Key"),
    x_byok_provider: Optional[str] = Header(default=None, alias="X-BYOK-Provider"),
    x_key_source: str = Header(default="platform", alias="X-Key-Source"),
    x_byok_base_url: Optional[str] = Header(default=None, alias="X-BYOK-Base-URL"),
):
    """Send a message to a chat session.

    Args:
        session_id: Existing session ID (or empty to create new)
        user_id: The user ID
        message: The message text
        tenant_id: Optional tenant ID; when set, chat context includes functions from the orchestrator
        api_key: Validated API key with chat:write scope
        x_byok_key: Optional BYOK API key from Go proxy
        x_byok_provider: Optional BYOK provider name from Go proxy
        x_key_source: Key source indicator ("byok", "token-plan", or "platform")
        x_byok_base_url: Optional BYOK base URL override (for MiMo Token Plan)

    Returns:
        ChatMessageResponse with the assistant's reply
    """
    try:
        sid = payload.session_id if payload is not None else (session_id or "")
        uid = payload.user_id if payload is not None else (user_id or "")
        msg = payload.message if payload is not None else (message or "")
        tid = payload.tenant_id if payload is not None else tenant_id
        ctx = payload.context if payload is not None else None

        if not msg:
            raise HTTPException(
                status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
                detail="message is required",
            )
        if not uid:
            uid = "unknown"

        # Check if BYOK key is provided — use it directly for LLM completion
        byok_key = x_byok_key if x_key_source in ("byok", "token-plan") else None
        byok_provider = x_byok_provider if x_key_source in ("byok", "token-plan") else None

        if byok_key and byok_provider:
            result = await _handle_byok_chat(
                session_id=sid,
                user_id=uid,
                message=msg,
                tenant_id=tid,
                byok_key=byok_key,
                byok_provider=byok_provider,
                base_url=x_byok_base_url,
                context=ctx,
            )
        else:
            # No BYOK key — try the chat manager (requires platform API keys)
            chat_manager = get_chat_manager()
            try:
                result = await asyncio.wait_for(
                    chat_manager.process_message(
                        session_id=sid,
                        user_id=uid,
                        message=msg,
                        tenant_id=tid,
                    ),
                    timeout=60.0,
                )
            except asyncio.TimeoutError:
                raise HTTPException(
                    status_code=status.HTTP_504_GATEWAY_TIMEOUT,
                    detail="Chat request timed out. Ensure an AI provider key is configured.",
                )

        return ChatMessageResponse(
            session_id=result["session_id"],
            message=result["message"],
            intent=ChatIntent(result.get("intent", "general")),
            confidence=result.get("confidence", 0.0),
            thinking_content=result.get("thinking_content"),
            thinking_tokens=result.get("thinking_tokens", 0),
        )
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Chat message failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to process chat message",
        )


async def _handle_byok_chat(
    session_id: str,
    user_id: str,
    message: str,
    tenant_id: Optional[str],
    byok_key: str,
    byok_provider: str,
    base_url: Optional[str],
    context: Optional[dict],
) -> dict:
    """Handle a chat message using a BYOK provider key directly."""
    from ..providers.manager import get_provider_manager
    from ..models.schemas import ChatMessage as SchemaChatMessage, MessageRole

    provider_manager = get_provider_manager()
    provider = provider_manager.get_provider_for_request(byok_provider, byok_key=byok_key, base_url=base_url)

    if not provider:
        raise HTTPException(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
            detail=f"Provider '{byok_provider}' not available",
        )

    # Build system prompt from context
    system_content = "You are a helpful AI assistant."
    thinking_config = None
    if context:
        if context.get("system_prompt"):
            system_content = context["system_prompt"]
        agent_name = context.get("agent_name", "")
        if agent_name:
            logger.info(f"BYOK chat for agent: {agent_name}")
        thinking_mode = context.get("thinking_mode", "off")
        if thinking_mode and thinking_mode != "off":
            budget = int(context.get("thinking_budget", "10000"))
            thinking_config = ThinkingConfig(mode=thinking_mode, budget_tokens=budget)

    messages = [
        ChatMessage(role=MessageRole.SYSTEM, content=system_content),
        ChatMessage(role=MessageRole.USER, content=message),
    ]

    try:
        response = await asyncio.wait_for(
            provider.complete(
                messages=messages,
                temperature=0.7,
                max_tokens=1024,
                thinking=thinking_config,
            ),
            timeout=90.0,
        )
        result = {
            "session_id": session_id,
            "message": response.content,
            "intent": "general",
            "confidence": 0.5,
        }
        if response.thinking_content:
            result["thinking_content"] = response.thinking_content
            result["thinking_tokens"] = response.thinking_tokens
        return result
    except asyncio.TimeoutError:
        logger.error(f"BYOK LLM completion timed out for provider {byok_provider}")
        return {
            "session_id": session_id,
            "message": "The AI provider took too long to respond. Please try again.",
            "intent": "general",
            "confidence": 0.0,
        }
    except Exception as e:
        logger.error(f"BYOK LLM completion failed: {e}")
        return {
            "session_id": session_id,
            "message": "I apologize, but I encountered an error processing your request. Please try again.",
            "intent": "general",
            "confidence": 0.0,
        }


@router.get("/api/chat/sessions", response_model=List[ChatSessionResponse])
async def list_chat_sessions(
    user_id: str = Query(..., description="User ID"),
    limit: int = Query(10, ge=1, le=50),
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.CHAT_READ)),
):
    """List chat sessions for a user.

    Args:
        user_id: The user ID
        limit: Maximum number of sessions
        api_key: Validated API key with chat:read scope

    Returns:
        List of ChatSessionResponse
    """
    try:
        chat_manager = get_chat_manager()
        sessions = await chat_manager.list_user_sessions(user_id, limit)
        return [
            ChatSessionResponse(
                session_id=s.session_id,
                user_id=s.user_id,
                created_at=s.created_at,
                message_count=len(s.messages),
            )
            for s in sessions
        ]
    except Exception as e:
        logger.error(f"Failed to list sessions: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to list sessions",
        )


@router.get("/api/chat/sessions/{session_id}", response_model=ChatHistoryResponse)
async def get_chat_session_history(
    session_id: str,
    api_key: APIKeyInfo = Depends(require_api_key_with_scope(KeyScope.CHAT_READ)),
):
    """Get chat session history.

    Args:
        session_id: The session ID
        api_key: Validated API key with chat:read scope

    Returns:
        ChatHistoryResponse with messages
    """
    try:
        chat_manager = get_chat_manager()
        session = await chat_manager.get_session(session_id)
        if not session:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="Session not found",
            )
        return ChatHistoryResponse(
            session_id=session.session_id,
            messages=session.messages,
            created_at=session.created_at,
        )
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Failed to get session history: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to get session history",
        )
