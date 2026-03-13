"""Intent classifier for chat service.

Simple rule-based classifier to detect user intent from natural language queries.
"""

import logging
import re
from enum import Enum
from typing import Optional

logger = logging.getLogger(__name__)


class ChatIntent(str, Enum):
    """Chat intent types."""
    EXPLAIN = "explain_intent"  # "why", "reason", "cause" - Explain latency causes
    QUERY = "query_intent"      # "show", "list", "find" - Query and display functions
    DEBUG = "debugging_intent"  # "fix", "error", "bug" - Debugging assistance
    OPTIMIZE = "optimization_intent"  # "optimize", "improve", "cost" - Cost/performance
    HELP = "help_intent"        # "help", "what" - General help
    UNKNOWN = "unknown_intent"  # Default fallback


class IntentClassifier:
    """Rule-based intent classifier for chat queries."""

    # Keyword patterns for each intent
    INTENT_PATTERNS = {
        ChatIntent.EXPLAIN: [
            r"\bwhy\b",
            r"\breason\b",
            r"\bcause\b",
            r"\bexplain\b",
            r"\bhow come\b",
            r"\bwhat caused\b",
        ],
        ChatIntent.QUERY: [
            r"\bshow\b",
            r"\blist\b",
            r"\bfind\b",
            r"\bget\b",
            r"\bdisplay\b",
            r"\bview\b",
            r"\bwhich\b",
            r"\bwhere\b",
            r"\ball\b",
        ],
        ChatIntent.DEBUG: [
            r"\bfix\b",
            r"\berror\b",
            r"\bbug\b",
            r"\bbroken\b",
            r"\bfailed\b",
            r"\bfailure\b",
            r"\bwrong\b",
            r"\bissue\b",
            r"\bproblem\b",
            r"\btroubleshoot\b",
            r"\bdebug\b",
        ],
        ChatIntent.OPTIMIZE: [
            r"\boptimize\b",
            r"\bimprove\b",
            r"\bcost\b",
            r"\bperformance\b",
            r"\bfaster\b",
            r"\bcheaper\b",
            r"\befficient\b",
            r"\brecommend\b",
            r"\btips\b",
            r"\bsuggest\b",
        ],
        ChatIntent.HELP: [
            r"\bhelp\b",
            r"\bwhat\b",
            r"\bhow\b",
            r"\bwhat can\b",
            r"\bwhat can you\b",
            r"\bwhat can i\b",
        ],
    }

    def __init__(self):
        """Initialize the intent classifier with compiled regex patterns."""
        self._compiled_patterns: dict[ChatIntent, list[re.Pattern]] = {}
        for intent, patterns in self.INTENT_PATTERNS.items():
            self._compiled_patterns[intent] = [re.compile(p, re.IGNORECASE) for p in patterns]

    def classify(self, message: str) -> tuple[ChatIntent, float]:
        """Classify the user's message intent.

        Args:
            message: The user's message

        Returns:
            Tuple of (intent, confidence_score)
        """
        if not message or not message.strip():
            return ChatIntent.UNKNOWN, 0.0

        message_lower = message.lower()
        scores: dict[ChatIntent, float] = {}

        # Calculate match score for each intent
        for intent, patterns in self._compiled_patterns.items():
            score = 0.0
            for pattern in patterns:
                if pattern.search(message_lower):
                    score += 1.0
            if score > 0:
                # Normalize by number of patterns for this intent
                scores[intent] = score / len(patterns)

        if not scores:
            return ChatIntent.UNKNOWN, 0.0

        # Return the intent with highest score
        best_intent = max(scores, key=scores.get)
        confidence = scores[best_intent]

        # Lower confidence threshold for ambiguous queries
        if confidence < 0.3:
            return ChatIntent.UNKNOWN, confidence

        logger.debug(f"Classified message as {best_intent.value} with confidence {confidence}")
        return best_intent, confidence

    def get_intent_keywords(self, intent: ChatIntent) -> list[str]:
        """Get keywords associated with an intent for prompt building.

        Args:
            intent: The intent to get keywords for

        Returns:
            List of keywords associated with the intent
        """
        keywords_map = {
            ChatIntent.EXPLAIN: ["latency", "slow", "performance", "delay", "bottleneck"],
            ChatIntent.QUERY: ["functions", "deployments", "errors", "metrics", "status"],
            ChatIntent.DEBUG: ["error", "exception", "traceback", "failed", "stack"],
            ChatIntent.OPTIMIZE: ["optimize", "improve", "cost", "performance", "memory"],
            ChatIntent.HELP: ["commands", "features", "capabilities", "api"],
            ChatIntent.UNKNOWN: [],
        }
        return keywords_map.get(intent, [])


# Global instance
_intent_classifier: Optional[IntentClassifier] = None


def get_intent_classifier() -> IntentClassifier:
    """Get the global intent classifier instance.

    Returns:
        The IntentClassifier instance
    """
    global _intent_classifier
    if _intent_classifier is None:
        _intent_classifier = IntentClassifier()
    return _intent_classifier
