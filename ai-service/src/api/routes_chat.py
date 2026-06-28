"""Chat endpoints for the developer experience layer."""

import logging
from typing import Optional, List

from fastapi import APIRouter, HTTPException, Query, status, Body, Depends
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
)
from ..services.chat import get_chat_manager

logger = logging.getLogger(__name__)

router = APIRouter()


class ChatSendBody(BaseModel):
    session_id: str
    user_id: str
    message: str
    tenant_id: Optional[str] = None


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
):
    """Send a message to a chat session.

    Args:
        session_id: Existing session ID (or empty to create new)
        user_id: The user ID
        message: The message text
        tenant_id: Optional tenant ID; when set, chat context includes functions from the orchestrator
        api_key: Validated API key with chat:write scope

    Returns:
        ChatMessageResponse with the assistant's reply
    """
    try:
        sid = payload.session_id if payload is not None else (session_id or "")
        uid = payload.user_id if payload is not None else (user_id or "")
        msg = payload.message if payload is not None else (message or "")
        tid = payload.tenant_id if payload is not None else tenant_id

        if not msg:
            raise HTTPException(
                status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
                detail="message is required",
            )
        if not uid:
            uid = "unknown"

        chat_manager = get_chat_manager()
        result = await chat_manager.process_message(
            session_id=sid,
            user_id=uid,
            message=msg,
            tenant_id=tid,
        )
        return ChatMessageResponse(
            session_id=result["session_id"],
            message=result["message"],
            intent=ChatIntent(result["intent"]),
            confidence=result["confidence"],
        )
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Chat message failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to process chat message",
        )


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
