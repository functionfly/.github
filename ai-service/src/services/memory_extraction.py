"""Team Memory Extraction Service for FlyMind.

This service analyzes conversations and extracts structured team memories
(decisions, preferences, processes, client context) using LLM analysis.
"""

import json
import logging
from typing import List, Dict, Any, Optional
from dataclasses import dataclass
from datetime import datetime

from ..providers.manager import ProviderManager, get_provider_manager
from ..providers.base import AIProvider


logger = logging.getLogger(__name__)


@dataclass
class MemoryExtraction:
    """A single extracted memory from a conversation."""
    type: str  # decision, preference, process, client_context
    category: Optional[str]
    summary: str
    content: Dict[str, Any]
    confidence: float
    rationale: str


@dataclass
class ExtractionResult:
    """Result of analyzing a conversation for memories."""
    memories: List[MemoryExtraction]
    confidence: float
    tokens_used: int
    model: str


# Extraction prompt template
EXTRACTION_SYSTEM_PROMPT = """You are a memory extraction specialist for team conversations.
Your task is to analyze conversations and extract structured memories that are valuable for the team.

Extract these types of memories:
1. **decision** - Important decisions made by the team, including rationale and alternatives considered
2. **preference** - Preferences expressed (communication styles, working hours, tools, etc.)
3. **process** - Workflows, procedures, or recurring processes mentioned
4. **client_context** - Information about clients/customers (names, requirements, preferences, important dates)

Rules:
- Only extract high-confidence facts (confidence >= 0.7)
- Prioritize: client preferences > team decisions > processes > general context
- Include specific details: names, dates, values, tools
- Mark preferences with priority 8-10 if they seem critical
- For decisions, include who made it and when
- If you see conflicting information, note it in rationale but extract the most recent
- DO NOT extract: casual chat, greetings, technical troubleshooting steps, temporary workarounds
- DO extract: "We decided to...", "The client prefers...", "Our process is...", "For Acme Corp..."

Respond ONLY with valid JSON in this exact format:
{
  "memories": [
    {
      "type": "decision|preference|process|client_context",
      "category": "optional category like 'client:acme-corp' or 'process:onboarding'",
      "summary": "concise human-readable summary (max 100 chars)",
      "content": {
        // For 'decision': { "title": "...", "rationale": "...", "decision_maker": "...", "alternatives": [...], "date": "..." }
        // For 'preference': { "subject": "...", "value": "...", "context": "...", "stakeholder": "...", "priority": 1-10 }
        // For 'process': { "name": "...", "steps": [...], "owner": "...", "frequency": "...", "tools": [...] }
        // For 'client_context': { "client_id": "...", "client_name": "...", "industry": "...", "preferences": {...}, "notes": "..." }
      },
      "confidence": 0.0-1.0,
      "rationale": "why this is important and what it means for the team"
    }
  ]
}"""


class MemoryExtractionService:
    """Service for extracting team memories from conversations."""

    def __init__(self, provider_manager: Optional[ProviderManager] = None):
        self.provider_manager = provider_manager or get_provider_manager()
        self._provider: Optional[AIProvider] = None

    async def _get_provider(self) -> AIProvider:
        """Get the best available provider for extraction.
        
        Uses gpt-4o-mini for optimal cost/quality balance (2026 pricing).
        Estimated cost: ~$0.00015 per 1K input tokens, ~$0.0006 per 1K output tokens.
        Typical extraction: ~2000 input + ~500 output tokens = ~$0.0006 per conversation.
        """
        if self._provider is None:
            # Prefer faster/cheaper models for extraction
            provider_manager = self.provider_manager
            # Try to get gpt-4o-mini first (good balance of quality and cost)
            try:
                self._provider = await provider_manager.get_provider("openai")
            except Exception:
                # Fallback to default
                self._provider = await provider_manager.get_default_provider()
        return self._provider

    async def analyze_conversation(
        self,
        transcript: str,
        context: Optional[Dict[str, Any]] = None,
    ) -> ExtractionResult:
        """Analyze a conversation transcript and extract memories.

        Args:
            transcript: The conversation transcript to analyze
            context: Optional context (team_id, conversation_id, participants, etc.)

        Returns:
            ExtractionResult containing extracted memories and metadata
        """
        provider = await self._get_provider()

        # Build the extraction prompt
        messages = [
            {"role": "system", "content": EXTRACTION_SYSTEM_PROMPT},
            {
                "role": "user",
                "content": f"Analyze the following team conversation and extract structured memories:\n\nCONVERSATION:\n{transcript}\n\nExtract memories in the specified JSON format."
            }
        ]

        try:
            # Call the provider
            response = await provider.chat_completion(
                messages=messages,
                model="gpt-4o-mini",  # Fast and cost-effective
                temperature=0.3,  # Low temperature for consistent extractions
                max_tokens=2000,
                json_mode=True,  # Request JSON output if supported
            )

            content = response.get("content", "")
            tokens_used = response.get("tokens_used", 0)
            model = response.get("model", "unknown")

            # Parse the JSON response
            try:
                # Handle potential markdown code blocks
                if "```json" in content:
                    content = content.split("```json")[1].split("```")[0].strip()
                elif "```" in content:
                    content = content.split("```")[1].split("```")[0].strip()

                parsed = json.loads(content)
                memories_data = parsed.get("memories", [])
            except json.JSONDecodeError as e:
                logger.warning(f"Failed to parse extraction response as JSON: {e}")
                logger.debug(f"Raw response: {content[:500]}...")
                return ExtractionResult(
                    memories=[],
                    confidence=0.0,
                    tokens_used=tokens_used,
                    model=model,
                )

            # Parse memories
            memories = []
            for mem_data in memories_data:
                try:
                    memory = MemoryExtraction(
                        type=mem_data.get("type", "preference"),
                        category=mem_data.get("category"),
                        summary=mem_data.get("summary", "Untitled"),
                        content=mem_data.get("content", {}),
                        confidence=float(mem_data.get("confidence", 0.7)),
                        rationale=mem_data.get("rationale", ""),
                    )
                    # Only include high-confidence memories
                    if memory.confidence >= 0.7:
                        memories.append(memory)
                except Exception as e:
                    logger.warning(f"Failed to parse memory extraction: {e}")
                    continue

            # Calculate overall confidence
            avg_confidence = sum(m.confidence for m in memories) / len(memories) if memories else 0.0

            return ExtractionResult(
                memories=memories,
                confidence=avg_confidence,
                tokens_used=tokens_used,
                model=model,
            )

        except Exception as e:
            logger.error(f"Memory extraction failed: {e}")
            return ExtractionResult(
                memories=[],
                confidence=0.0,
                tokens_used=0,
                model="error",
            )

    async def batch_analyze(
        self,
        transcripts: List[str],
        context: Optional[Dict[str, Any]] = None,
    ) -> List[ExtractionResult]:
        """Analyze multiple conversations in batch.

        Args:
            transcripts: List of conversation transcripts
            context: Optional context

        Returns:
            List of ExtractionResults
        """
        results = []
        for transcript in transcripts:
            result = await self.analyze_conversation(transcript, context)
            results.append(result)
        return results


# Singleton instance
_extraction_service: Optional[MemoryExtractionService] = None


def get_memory_extraction_service() -> MemoryExtractionService:
    """Get the singleton memory extraction service instance."""
    global _extraction_service
    if _extraction_service is None:
        _extraction_service = MemoryExtractionService()
    return _extraction_service
