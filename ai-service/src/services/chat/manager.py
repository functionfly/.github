"""Chat manager for FlyMind AI Service.

Handles chat session management and message processing.
"""

import json
import logging
import uuid
from datetime import datetime, timedelta
from typing import Optional, Any, Dict, List

import redis.asyncio as redis

from ...config import settings
from ...models.schemas import ChatMessage, MessageRole
from ...providers.manager import get_provider_manager
from .intent_classifier import IntentClassifier, get_intent_classifier
from .context_builder import ContextBuilder, get_context_builder
from .rag import get_rag_index
from . import prompts

logger = logging.getLogger(__name__)

# Session configuration
MAX_MESSAGES_PER_SESSION = 10  # Keep last 10 messages for context window
SESSION_TTL_SECONDS = 3600  # 1 hour session expiry


class ChatSession:
    """Represents a chat session."""

    def __init__(
        self,
        session_id: str,
        user_id: str,
        created_at: datetime,
        messages: List[Dict[str, Any]],
    ):
        self.session_id = session_id
        self.user_id = user_id
        self.created_at = created_at
        self.messages = messages

    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary."""
        return {
            "session_id": self.session_id,
            "user_id": self.user_id,
            "created_at": self.created_at.isoformat(),
            "messages": self.messages,
        }

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "ChatSession":
        """Create from dictionary."""
        return cls(
            session_id=data["session_id"],
            user_id=data["user_id"],
            created_at=datetime.fromisoformat(data["created_at"]),
            messages=data.get("messages", []),
        )


class ChatManager:
    """Manages chat sessions and message processing."""

    def __init__(self):
        self._redis: Optional[redis.Redis] = None
        self._intent_classifier = get_intent_classifier()
        self._context_builder = get_context_builder()

    async def get_redis(self) -> Optional[redis.Redis]:
        """Get Redis connection."""
        if self._redis is None:
            try:
                self._redis = redis.from_url(
                    settings.redis_url,
                    encoding="utf-8",
                    decode_responses=True,
                )
                await self._redis.ping()
            except Exception as e:
                logger.warning(f"Failed to connect to Redis: {e}")
                self._redis = None
        return self._redis

    def _get_session_key(self, session_id: str) -> str:
        """Get Redis key for session."""
        return f"chat:session:{session_id}"

    def _get_user_sessions_key(self, user_id: str) -> str:
        """Get Redis key for user's session list."""
        return f"chat:user:{user_id}:sessions"

    async def create_session(self, user_id: str) -> ChatSession:
        """Create a new chat session.

        Args:
            user_id: The user ID

        Returns:
            New ChatSession
        """
        session_id = str(uuid.uuid4())
        session = ChatSession(
            session_id=session_id,
            user_id=user_id,
            created_at=datetime.utcnow(),
            messages=[],
        )

        # Save to Redis
        redis_client = await self.get_redis()
        if redis_client:
            try:
                await redis_client.setex(
                    self._get_session_key(session_id),
                    SESSION_TTL_SECONDS,
                    json.dumps(session.to_dict()),
                )
                # Add to user's session list
                await redis_client.sadd(
                    self._get_user_sessions_key(user_id),
                    session_id,
                )
            except Exception as e:
                logger.error(f"Failed to save session: {e}")

        logger.info(f"Created chat session {session_id} for user {user_id}")
        return session

    async def get_session(self, session_id: str) -> Optional[ChatSession]:
        """Get a chat session by ID.

        Args:
            session_id: The session ID

        Returns:
            ChatSession or None if not found
        """
        redis_client = await self.get_redis()
        if not redis_client:
            return None

        try:
            data = await redis_client.get(self._get_session_key(session_id))
            if data:
                session_dict = json.loads(data)
                # Update TTL on access
                await redis_client.expire(
                    self._get_session_key(session_id),
                    SESSION_TTL_SECONDS,
                )
                return ChatSession.from_dict(session_dict)
        except Exception as e:
            logger.error(f"Failed to get session: {e}")

        return None

    async def list_user_sessions(
        self,
        user_id: str,
        limit: int = 10,
    ) -> List[ChatSession]:
        """List user's chat sessions.

        Args:
            user_id: The user ID
            limit: Maximum number of sessions to return

        Returns:
            List of ChatSession objects
        """
        redis_client = await self.get_redis()
        if not redis_client:
            return []

        try:
            session_ids = await redis_client.smembers(
                self._get_user_sessions_key(user_id),
            )

            sessions = []
            for session_id in list(session_ids)[:limit]:
                session = await self.get_session(session_id)
                if session:
                    sessions.append(session)

            # Sort by creation time, newest first
            sessions.sort(key=lambda s: s.created_at, reverse=True)
            return sessions
        except Exception as e:
            logger.error(f"Failed to list sessions: {e}")
            return []

    async def add_message(
        self,
        session_id: str,
        role: str,
        content: str,
        intent: Optional[str] = None,
        confidence: Optional[float] = None,
    ) -> Optional[ChatSession]:
        """Add a message to a session.

        Args:
            session_id: The session ID
            role: Message role (user/assistant)
            content: Message content
            intent: Detected intent (for user messages)
            confidence: Intent confidence score

        Returns:
            Updated ChatSession or None
        """
        session = await self.get_session(session_id)
        if not session:
            logger.warning(f"Session {session_id} not found")
            return None

        # Create message
        message = {
            "role": role,
            "content": content,
            "timestamp": datetime.utcnow().isoformat(),
        }
        if intent:
            message["intent"] = intent
        if confidence is not None:
            message["confidence"] = confidence

        # Add to messages
        session.messages.append(message)

        # Keep only last MAX_MESSAGES_PER_SESSION
        if len(session.messages) > MAX_MESSAGES_PER_SESSION:
            session.messages = session.messages[-MAX_MESSAGES_PER_SESSION:]

        # Save to Redis
        redis_client = await self.get_redis()
        if redis_client:
            try:
                await redis_client.setex(
                    self._get_session_key(session_id),
                    SESSION_TTL_SECONDS,
                    json.dumps(session.to_dict()),
                )
            except Exception as e:
                logger.error(f"Failed to save session: {e}")

        return session

    async def process_message(
        self,
        session_id: str,
        user_id: str,
        message: str,
        tenant_id: Optional[str] = None,
    ) -> Dict[str, Any]:
        """Process a user message and generate a response.

        Args:
            session_id: The session ID
            user_id: The user ID
            message: The user's message
            tenant_id: Optional tenant ID for orchestrator-backed context (e.g. deployed functions)

        Returns:
            Dict with response and metadata
        """
        # Get or create session
        session = await self.get_session(session_id)
        if not session:
            session = await self.create_session(user_id)
            session_id = session.session_id

        # Classify intent
        intent, confidence = self._intent_classifier.classify(message)
        logger.info(f"Classified intent: {intent.value} (confidence: {confidence})")

        # Add user message to session
        await self.add_message(
            session_id=session_id,
            role="user",
            content=message,
            intent=intent.value,
            confidence=confidence,
        )

        # Build context
        context = await self._context_builder.build_context(
            user_id=user_id,
            intent=intent.value,
            query=message,
            tenant_id=tenant_id,
        )

        # RAG: retrieve relevant docs excerpts and append to context.
        try:
            rag_block = await get_rag_index().build_context_block(message)
            if rag_block:
                context = f"{context}\n\n{rag_block}"
        except Exception as e:
            logger.warning(f"RAG retrieval failed (continuing without it): {e}")

        # Get system prompt
        system_prompt = prompts.get_system_prompt(intent, context, message)

        # Build messages for LLM (providers expect ChatMessage, not raw dicts)
        messages_for_llm: List[ChatMessage] = [
            ChatMessage(role=MessageRole.SYSTEM, content=system_prompt)
        ]

        # Add recent conversation history
        for msg in session.messages[-MAX_MESSAGES_PER_SESSION:]:
            # Skip system messages we already included
            if msg["role"] != "system":
                messages_for_llm.append(
                    ChatMessage(
                        role=MessageRole(msg["role"]),
                        content=msg["content"],
                    )
                )

        # Get response from LLM
        try:
            provider_manager = get_provider_manager()
            provider = provider_manager.get_provider_for_chat()

            response = await provider.complete(
                messages=messages_for_llm,
                temperature=0.7,
                max_tokens=1024,
            )

            assistant_message = response.content
        except Exception as e:
            logger.error(f"LLM completion failed: {e}")
            assistant_message = "I apologize, but I encountered an error processing your request. Please try again."

        # Add assistant message to session
        await self.add_message(
            session_id=session_id,
            role="assistant",
            content=assistant_message,
        )

        return {
            "session_id": session_id,
            "message": assistant_message,
            "intent": intent.value,
            "confidence": confidence,
        }

    async def delete_session(self, session_id: str) -> bool:
        """Delete a chat session.

        Args:
            session_id: The session ID

        Returns:
            True if deleted successfully
        """
        redis_client = await self.get_redis()
        if not redis_client:
            return False

        try:
            session = await self.get_session(session_id)
            if session:
                # Delete from user's session set
                await redis_client.srem(
                    self._get_user_sessions_key(session.user_id),
                    session_id,
                )
                # Delete session data
                await redis_client.delete(self._get_session_key(session_id))
                logger.info(f"Deleted session {session_id}")
                return True
        except Exception as e:
            logger.error(f"Failed to delete session: {e}")

        return False

    async def close(self):
        """Close Redis connection."""
        if self._redis:
            await self._redis.close()
            self._redis = None


# Global instance
_chat_manager: Optional[ChatManager] = None


def get_chat_manager() -> ChatManager:
    """Get the global chat manager instance.

    Returns:
        The ChatManager instance
    """
    global _chat_manager
    if _chat_manager is None:
        _chat_manager = ChatManager()
    return _chat_manager
