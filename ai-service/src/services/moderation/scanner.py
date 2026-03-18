"""Content scanner for policy violations.

This module provides content scanning capabilities using:
- Regex patterns for PII and secrets
- Optional OpenAI Moderation API (omni-moderation-latest, 2026 recommended) for
  toxic/hate/violence/sexual/self-harm/illicit when OPENAI_API_KEY is set
- Optional self-hosted ML (Detoxify) when OpenAI is not used
- Keyword fallback when no API/ML is available
"""

import asyncio
import re
import hashlib
import logging
from typing import List, Optional, Dict, Any, Tuple

from ...config import settings
from .results import Violation, ModerationCategory

logger = logging.getLogger(__name__)

# Max input length for ML model (Detoxify/BERT ~512 subwords; ~2000 chars safe)
ML_MODERATION_MAX_CHARS = 2000

# Score threshold above which we emit a violation (0.0-1.0)
ML_MODERATION_THRESHOLD = 0.5

# =============================================================================
# Pattern Definitions
# =============================================================================

# PII Patterns
PII_PATTERNS = {
    # Email addresses
    "email": re.compile(
        r'\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b',
        re.IGNORECASE
    ),
    # Phone numbers (US format)
    "phone_us": re.compile(
        r'\b(?:\+1[-.\s]?)?(?:\([0-9]{3}\)|[0-9]{3})[-.\s]?[0-9]{3}[-.\s]?[0-9]{4}\b'
    ),
    # Social Security Numbers
    "ssn": re.compile(
        r'\b[0-9]{3}[-\s]?[0-9]{2}[-\s]?[0-9]{4}\b'
    ),
    # Credit card numbers (basic pattern)
    "credit_card": re.compile(
        r'\b(?:4[0-9]{12}(?:[0-9]{3})?|5[1-5][0-9]{14}|3[47][0-9]{13}|6(?:011|5[0-9]{2})[0-9]{12})\b'
    ),
    # IP addresses
    "ip_address": re.compile(
        r'\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b'
    ),
    # Physical addresses (basic pattern)
    "address": re.compile(
        r'\b\d{1,5}\s+[\w\s]+(?:street|st|avenue|ave|road|rd|boulevard|blvd|drive|dr|lane|ln|court|ct|way|place|pl)\b',
        re.IGNORECASE
    ),
}

# Secret/API Key Patterns
SECRET_PATTERNS = {
    # Generic API key patterns
    "api_key_generic": re.compile(
        r'\b(?:api[_-]?key|apikey|api_secret|apisecret)[=:\s]+["\']?[A-Za-z0-9_\-]{16,}["\']?',
        re.IGNORECASE
    ),
    # AWS Access Key
    "aws_access_key": re.compile(
        r'\b(AKIA|ABIA|ACCA|ASIA)[0-9A-Z]{16}\b'
    ),
    # AWS Secret Key
    "aws_secret_key": re.compile(
        r'\b[A-Za-z0-9/+=]{40}\b',
        re.IGNORECASE
    ),
    # GitHub Token
    "github_token": re.compile(
        r'\b(?:ghp|gho|ghu|ghs|ghr)_[A-Za-z0-9_]{36,}\b'
    ),
    # Generic JWT tokens
    "jwt": re.compile(
        r'\beyJ[A-Za-z0-9_-]+\.eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\b'
    ),
    # Private key patterns
    "private_key": re.compile(
        r'-----BEGIN\s+(?:RSA\s+)?(?:DSA\s+)?(?:EC\s+)?PRIVATE\s+KEY-----'
    ),
    # Database connection strings
    "db_connection": re.compile(
        r'(?:mongodb|postgresql|mysql|redis)://[^\s]+',
        re.IGNORECASE
    ),
    # Bearer tokens
    "bearer_token": re.compile(
        r'\bBearer\s+[A-Za-z0-9\-._~+/]+=*\b',
        re.IGNORECASE
    ),
    # Basic auth
    "basic_auth": re.compile(
        r'\bBasic\s+[A-Za-z0-9+/]+=*\b',
        re.IGNORECASE
    ),
    # Slack tokens
    "slack_token": re.compile(
        r'\bxox[baprs]-[0-9]{10,}-[0-9]{10,}-[a-zA-Z0-9]{24,}\b'
    ),
    # Stripe keys
    "stripe_key": re.compile(
        r'\b(?:sk|pk)_(?:test|live)_[0-9a-zA-Z]{24,}\b'
    ),
    # OpenAI API key
    "openai_key": re.compile(
        r'\bsk-[A-Za-z0-9]{20,}\b'
    ),
    # Anthropic API key
    "anthropic_key": re.compile(
        r'\b(?:sk-ant|cln)-[A-Za-z0-9_-]{20,}\b',
        re.IGNORECASE
    ),
}

# Toxic / policy-violation keywords. Used only when neither OpenAI Moderation API
# nor Detoxify is available. Prefer moderation_provider=openai (or auto with
# OPENAI_API_KEY) for production.
TOXIC_KEYWORDS = [
    "malware",
    "phishing",
    "ransomware",
    "exploit",
    "hack",
    "bypass",
    "inject",
    "exfiltrate",
    "keylogger",
    "trojan",
    "backdoor",
    "credential theft",
    "ddos",
    "sql injection",
    "xss",
    "csrf",
]

# Fallback keyword lists when ML (Detoxify) is not installed or fails
HATE_SPEECH_KEYWORDS = [
    "hate", "hatred", "bigot", "racist", "sexist", "nazi", "supremacist",
    "genocide", "ethnic cleansing", "dehumaniz",
]

VIOLENCE_KEYWORDS = [
    "kill you", "murder", "attack you", "beat you", "bomb", "shoot you",
    "stab", "threaten", "violence", "assault",
]


# -----------------------------------------------------------------------------
# Self-hosted ML moderation (Detoxify) - optional
# -----------------------------------------------------------------------------

_detoxify_model: Any = None
_detoxify_loaded: bool = False
_detoxify_error: Optional[str] = None


def _load_detoxify_model() -> Tuple[Any, Optional[str]]:
    """Load Detoxify model once (sync, run in thread). Returns (model, error)."""
    global _detoxify_model, _detoxify_loaded, _detoxify_error
    if _detoxify_loaded:
        return _detoxify_model, _detoxify_error
    _detoxify_loaded = True
    try:
        from detoxify import Detoxify
        # original-small is lighter; use 'original' for best accuracy
        _detoxify_model = Detoxify("original-small")
        logger.info("Detoxify model loaded for self-hosted moderation")
        return _detoxify_model, None
    except Exception as e:
        _detoxify_error = str(e)
        logger.debug("Detoxify not available for moderation: %s", _detoxify_error)
        return None, _detoxify_error


def _run_detoxify_sync(text: str) -> Dict[str, float]:
    """Run Detoxify predict (sync). Call from thread."""
    model, err = _load_detoxify_model()
    if model is None or not text.strip():
        return {}
    text_trimmed = text[:ML_MODERATION_MAX_CHARS].strip()
    if not text_trimmed:
        return {}
    try:
        return model.predict(text_trimmed)
    except Exception as e:
        logger.warning("Detoxify predict failed: %s", e)
        return {}


def _scores_to_violations(scores: Dict[str, float]) -> List[Violation]:
    """Map Detoxify scores to our Violation list. Keys: toxic, severe_toxic, obscene, threat, insult, identity_attack."""
    violations: List[Violation] = []
    # identity_attack, insult -> hate_speech
    hate_score = max(
        scores.get("identity_attack", 0),
        scores.get("insult", 0),
    )
    if hate_score >= ML_MODERATION_THRESHOLD:
        violations.append(Violation(
            category=ModerationCategory.HATE_SPEECH,
            severity="high",
            message="Potentially hateful or identity-attacking content (ML)",
            matched_pattern="ml_detoxify",
            location_start=None,
            location_end=None,
            confidence=round(hate_score, 2),
        ))
    # threat -> violence
    threat = scores.get("threat", 0)
    if threat >= ML_MODERATION_THRESHOLD:
        violations.append(Violation(
            category=ModerationCategory.VIOLENCE,
            severity="high",
            message="Potentially threatening or violent content (ML)",
            matched_pattern="ml_detoxify",
            location_start=None,
            location_end=None,
            confidence=round(threat, 2),
        ))
    # toxic, severe_toxic, obscene -> TOXIC category
    toxic_score = max(
        scores.get("toxic", 0),
        scores.get("severe_toxic", 0),
        scores.get("obscene", 0),
    )
    if toxic_score >= ML_MODERATION_THRESHOLD:
        violations.append(Violation(
            category=ModerationCategory.TOXIC,
            severity="high" if scores.get("severe_toxic", 0) >= ML_MODERATION_THRESHOLD else "medium",
            message="Potentially toxic or obscene content (ML)",
            matched_pattern="ml_detoxify",
            location_start=None,
            location_end=None,
            confidence=round(toxic_score, 2),
        ))
    return violations


# -----------------------------------------------------------------------------
# OpenAI Moderation API (omni-moderation-latest recommended for 2026)
# -----------------------------------------------------------------------------


def _openai_categories_to_violations(result: Any) -> List[Violation]:
    """Map OpenAI moderation result (first item) to our Violation list."""
    violations: List[Violation] = []
    if not result or not getattr(result, "results", None) or len(result.results) == 0:
        return violations

    r = result.results[0]
    categories = getattr(r, "categories", None)
    scores = getattr(r, "category_scores", None)
    if not categories:
        return violations

    def score_for(name: str) -> float:
        if not scores:
            return 0.0
        return getattr(scores, name, 0.0) or 0.0

    def flagged(name: str) -> bool:
        return getattr(categories, name, False) or False

    # Hate -> HATE_SPEECH
    if flagged("hate") or flagged("hate_threatening"):
        violations.append(Violation(
            category=ModerationCategory.HATE_SPEECH,
            severity="high" if flagged("hate_threatening") else "medium",
            message="Content flagged for hate (OpenAI Moderation)",
            matched_pattern="openai_moderation",
            location_start=None,
            location_end=None,
            confidence=round(max(score_for("hate"), score_for("hate_threatening")), 2),
        ))

    # Violence -> VIOLENCE
    if flagged("violence") or flagged("violence_graphic"):
        violations.append(Violation(
            category=ModerationCategory.VIOLENCE,
            severity="high" if flagged("violence_graphic") else "medium",
            message="Content flagged for violence (OpenAI Moderation)",
            matched_pattern="openai_moderation",
            location_start=None,
            location_end=None,
            confidence=round(max(score_for("violence"), score_for("violence_graphic")), 2),
        ))

    # Sexual -> SEXUAL
    if flagged("sexual") or flagged("sexual_minors"):
        violations.append(Violation(
            category=ModerationCategory.SEXUAL,
            severity="critical" if flagged("sexual_minors") else "high",
            message="Content flagged for sexual content (OpenAI Moderation)",
            matched_pattern="openai_moderation",
            location_start=None,
            location_end=None,
            confidence=round(max(score_for("sexual"), score_for("sexual_minors")), 2),
        ))

    # Harassment -> TOXIC
    if flagged("harassment") or flagged("harassment_threatening"):
        violations.append(Violation(
            category=ModerationCategory.TOXIC,
            severity="high" if flagged("harassment_threatening") else "medium",
            message="Content flagged for harassment (OpenAI Moderation)",
            matched_pattern="openai_moderation",
            location_start=None,
            location_end=None,
            confidence=round(max(score_for("harassment"), score_for("harassment_threatening")), 2),
        ))

    # Self-harm -> TOXIC (we don't have a dedicated category)
    if flagged("self_harm") or flagged("self_harm_intent") or flagged("self_harm_instructions"):
        violations.append(Violation(
            category=ModerationCategory.TOXIC,
            severity="high",
            message="Content flagged for self-harm (OpenAI Moderation)",
            matched_pattern="openai_moderation",
            location_start=None,
            location_end=None,
            confidence=round(max(
                score_for("self_harm"),
                score_for("self_harm_intent"),
                score_for("self_harm_instructions"),
            ), 2),
        ))

    # Illicit (optional on omni-moderation)
    if getattr(categories, "illicit", False) or getattr(categories, "illicit_violent", False):
        violations.append(Violation(
            category=ModerationCategory.TOXIC,
            severity="high" if getattr(categories, "illicit_violent", False) else "medium",
            message="Content flagged for illicit (OpenAI Moderation)",
            matched_pattern="openai_moderation",
            location_start=None,
            location_end=None,
            confidence=round(max(score_for("illicit"), score_for("illicit_violent")), 2),
        ))

    return violations


class ContentScanner:
    """Scanner for detecting policy violations in content."""

    def __init__(self):
        """Initialize the content scanner."""
        self._logger = logging.getLogger(__name__)
        self._scan_count = 0
        self._violation_count = 0

    async def scan(
        self,
        content: str,
        content_type: str = "text",
    ) -> List[Violation]:
        """Scan content for policy violations.

        Args:
            content: The content to scan
            content_type: Type of content (text, json, code)

        Returns:
            List of violations found
        """
        violations = []

        # Skip empty content
        if not content or not content.strip():
            return violations

        self._scan_count += 1

        # Scan for PII
        pii_violations = self._scan_pii(content)
        violations.extend(pii_violations)

        # Scan for secrets
        secret_violations = self._scan_secrets(content)
        violations.extend(secret_violations)

        # Toxic / hate / violence / sexual: OpenAI Moderation API (2026 preferred),
        # then Detoxify, then keyword fallback
        harm_violations = await self._run_harm_moderation(content)
        if harm_violations:
            violations.extend(harm_violations)
        else:
            toxic_violations = self._scan_toxic(content)
            violations.extend(toxic_violations)
            hate_violations = self._scan_hate_speech(content)
            violations.extend(hate_violations)
            violence_violations = self._scan_violence(content)
            violations.extend(violence_violations)

        # Scan for malware patterns
        malware_violations = self._scan_malware(content)
        violations.extend(malware_violations)

        # Scan for spam
        spam_violations = self._scan_spam(content)
        violations.extend(spam_violations)

        if violations:
            self._violation_count += len(violations)
            self._logger.info(
                f"Found {len(violations)} violations in content "
                f"(scan #{self._scan_count})"
            )

        return violations

    async def _run_harm_moderation(self, content: str) -> List[Violation]:
        """Run harm detection: OpenAI Moderation API (preferred in 2026), then Detoxify, else []."""
        if not content or not content.strip():
            return []
        provider = (settings.moderation_provider or "auto").lower()

        # Prefer OpenAI Moderation API when key is set and provider is openai or auto
        if provider in ("openai", "auto") and (settings.openai_api_key or "").strip():
            try:
                return await self._run_openai_moderation(content)
            except Exception as e:
                self._logger.debug("OpenAI moderation failed, falling back: %s", e)

        # Self-hosted Detoxify when provider is detoxify or auto
        if provider in ("detoxify", "auto"):
            return await self._run_ml_moderation(content)

        return []

    async def _run_openai_moderation(self, content: str) -> List[Violation]:
        """Call OpenAI Moderation API (omni-moderation-latest); return violations if any."""
        text = content.strip()[:ML_MODERATION_MAX_CHARS]
        if not text:
            return []
        try:
            from openai import AsyncOpenAI

            client = AsyncOpenAI(api_key=settings.openai_api_key)
            result = await client.moderations.create(
                input=text,
                model=settings.openai_moderation_model,
            )
            return _openai_categories_to_violations(result)
        except Exception as e:
            self._logger.warning("OpenAI moderation API error: %s", e)
            raise

    async def _run_ml_moderation(self, content: str) -> List[Violation]:
        """Run self-hosted Detoxify model in a thread; return violations if any."""
        if not content or len(content.strip()) == 0:
            return []
        try:
            loop = asyncio.get_event_loop()
            scores = await loop.run_in_executor(None, _run_detoxify_sync, content)
        except Exception as e:
            self._logger.debug("ML moderation skipped or failed: %s", e)
            return []
        if not scores:
            return []
        return _scores_to_violations(scores)

    def _scan_pii(self, content: str) -> List[Violation]:
        """Scan for PII in content.

        Args:
            content: Content to scan

        Returns:
            List of PII violations
        """
        violations = []

        for pattern_name, pattern in PII_PATTERNS.items():
            matches = pattern.finditer(content)
            for match in matches:
                # Calculate severity based on match context
                severity = self._calculate_severity(
                    ModerationCategory.PII,
                    pattern_name,
                    match.group()
                )

                violations.append(Violation(
                    category=ModerationCategory.PII,
                    severity=severity,
                    message=f"Potential {pattern_name.replace('_', ' ')} detected",
                    matched_pattern=pattern_name,
                    location_start=match.start(),
                    location_end=match.end(),
                    confidence=0.8,
                ))

        return violations

    def _scan_secrets(self, content: str) -> List[Violation]:
        """Scan for secrets and API keys in content.

        Args:
            content: Content to scan

        Returns:
            List of secret violations
        """
        violations = []

        for pattern_name, pattern in SECRET_PATTERNS.items():
            matches = pattern.finditer(content)
            for match in matches:
                # Always high severity for secrets
                severity = "high"

                violations.append(Violation(
                    category=ModerationCategory.SECRETS,
                    severity=severity,
                    message=f"Potential {pattern_name.replace('_', ' ')} detected",
                    matched_pattern=pattern_name,
                    location_start=match.start(),
                    location_end=match.end(),
                    confidence=0.9,
                ))

        return violations

    def _scan_toxic(self, content: str) -> List[Violation]:
        """Scan for toxic content in content.

        Args:
            content: Content to scan

        Returns:
            List of toxic content violations
        """
        violations = []
        content_lower = content.lower()

        for keyword in TOXIC_KEYWORDS:
            if keyword in content_lower:
                # Find the position
                pos = content_lower.find(keyword)

                violations.append(Violation(
                    category=ModerationCategory.TOXIC,
                    severity="medium",
                    message=f"Potentially toxic keyword detected: {keyword}",
                    matched_pattern=keyword,
                    location_start=pos,
                    location_end=pos + len(keyword),
                    confidence=0.6,
                ))

        return violations

    def _scan_malware(self, content: str) -> List[Violation]:
        """Scan for malware patterns in content.

        Args:
            content: Content to scan

        Returns:
            List of malware violations
        """
        violations = []

        # Common malware patterns
        malware_patterns = [
            (r'<script[^>]*>.*?<\/script>', "JavaScript injection"),
            (r'eval\s*\(', "Eval function usage"),
            (r'document\.cookie', "Cookie access"),
            (r'window\.location\s*=', "URL redirection"),
            (r'exec\s*\(', "Execute function usage"),
            (r'system\s*\(', "System call"),
            (r'shell_exec\s*\(', "Shell execution"),
            (r'passthru\s*\(', "Passthru call"),
            (r'base64_decode\s*\(', "Base64 decode"),
            (r'chr\s*\(\s*\d+\s*\)', "Char code obfuscation"),
        ]

        for pattern_str, description in malware_patterns:
            pattern = re.compile(pattern_str, re.IGNORECASE | re.DOTALL)
            matches = pattern.finditer(content)

            for match in matches:
                violations.append(Violation(
                    category=ModerationCategory.MALWARE,
                    severity="high",
                    message=f"Potential malware pattern: {description}",
                    matched_pattern=description,
                    location_start=match.start(),
                    location_end=match.end(),
                    confidence=0.7,
                ))

        return violations

    def _scan_hate_speech(self, content: str) -> List[Violation]:
        """Scan for hate speech (keyword fallback when ML not available)."""
        violations = []
        content_lower = content.lower()

        for keyword in HATE_SPEECH_KEYWORDS:
            if keyword in content_lower:
                pos = content_lower.find(keyword)

                violations.append(Violation(
                    category=ModerationCategory.HATE_SPEECH,
                    severity="high",
                    message=f"Potentially hateful content detected: {keyword}",
                    matched_pattern=keyword,
                    location_start=pos,
                    location_end=pos + len(keyword),
                    confidence=0.5,
                ))

        return violations

    def _scan_violence(self, content: str) -> List[Violation]:
        """Scan for violence (keyword fallback when ML not available)."""
        violations = []
        content_lower = content.lower()

        for keyword in VIOLENCE_KEYWORDS:
            if keyword in content_lower:
                pos = content_lower.find(keyword)

                violations.append(Violation(
                    category=ModerationCategory.VIOLENCE,
                    severity="medium",
                    message=f"Violence-related keyword detected: {keyword}",
                    matched_pattern=keyword,
                    location_start=pos,
                    location_end=pos + len(keyword),
                    confidence=0.5,
                ))

        return violations

    def _scan_spam(self, content: str) -> List[Violation]:
        """Scan for spam patterns in content.

        Args:
            content: Content to scan

        Returns:
            List of spam violations
        """
        violations = []

        # Spam patterns
        spam_patterns = [
            # Excessive caps
            (r'[A-Z]{10,}', "Excessive capitalization"),
            # Multiple exclamation marks
            (r'[!]{5,}', "Excessive exclamation marks"),
            # Suspicious URLs
            (r'https?://[^\s]*\.(?:xyz|tk|ml|ga|cf|gq|top|work|click|stream)[^\s]*',
             "Suspicious URL domain"),
            # Money-related spam
            (r'(?:free|money|win|prize|cash|bonus).*(?:money|cash|prize|win|free)',
             "Money-related spam"),
        ]

        for pattern_str, description in spam_patterns:
            pattern = re.compile(pattern_str, re.IGNORECASE)
            matches = pattern.finditer(content)

            for match in matches:
                violations.append(Violation(
                    category=ModerationCategory.SPAM,
                    severity="low",
                    message=f"Potential spam: {description}",
                    matched_pattern=description,
                    location_start=match.start(),
                    location_end=match.end(),
                    confidence=0.5,
                ))

        return violations

    def _calculate_severity(
        self,
        category: ModerationCategory,
        pattern_name: str,
        matched_text: str,
    ) -> str:
        """Calculate the severity of a violation.

        Args:
            category: The violation category
            pattern_name: The pattern that was matched
            matched_text: The matched text

        Returns:
            Severity level (low, medium, high, critical)
        """
        # Critical severity for certain patterns
        critical_patterns = ["ssn", "private_key", "aws_secret_key"]
        if pattern_name in critical_patterns:
            return "critical"

        # High severity for secrets
        if category == ModerationCategory.SECRETS:
            return "high"

        # Medium severity for PII
        if category == ModerationCategory.PII:
            return "medium"

        # Default to medium
        return "medium"

    def compute_content_hash(self, content: str) -> str:
        """Compute a hash of the content for deduplication.

        Args:
            content: Content to hash

        Returns:
            SHA256 hash of the content
        """
        return hashlib.sha256(content.encode()).hexdigest()

    @property
    def stats(self) -> Dict[str, int]:
        """Get scanner statistics.

        Returns:
            Dictionary with scan statistics
        """
        return {
            "total_scans": self._scan_count,
            "total_violations": self._violation_count,
        }


# Global scanner instance
_scanner: Optional[ContentScanner] = None


def get_content_scanner() -> ContentScanner:
    """Get the global content scanner instance.

    Returns:
        ContentScanner instance
    """
    global _scanner
    if _scanner is None:
        _scanner = ContentScanner()

    return _scanner
