"""Fix suggester for debugging service.

Generates fix suggestions with code examples.
"""

import logging
from typing import Optional, Dict, Any, List

from ...providers.manager import get_provider_manager

logger = logging.getLogger(__name__)


class FixSuggester:
    """Generates fix suggestions for identified errors."""

    def __init__(self):
        self._fix_templates = self._load_fix_templates()

    def _load_fix_templates(self) -> Dict[str, List[Dict[str, str]]]:
        """Load fix suggestion templates."""
        return {
            "timeout": [
                {
                    "title": "Increase Timeout",
                    "description": "Increase the function timeout in your configuration.",
                    "code": """# functionfly.jsonc
{
  "timeout": 60
}""",
                    "effort": "low",
                    "impact": "medium",
                },
                {
                    "title": "Add Caching",
                    "description": "Cache expensive API calls to reduce execution time.",
                    "code": """import json

# Simple in-memory cache
cache = {}

def handler(request):
    cache_key = json.dumps(request)

    if cache_key in cache:
        return cache[cache_key]

    # Your API call here
    result = call_external_api(request)

    cache[cache_key] = result
    return result""",
                    "effort": "medium",
                    "impact": "high",
                },
            ],
            "memory": [
                {
                    "title": "Process in Chunks",
                    "description": "Process large datasets in smaller chunks to reduce memory usage.",
                    "code": """def handler(request):
    items = request.get("items", [])
    batch_size = 100

    results = []
    for i in range(0, len(items), batch_size):
        batch = items[i:i + batch_size]
        results.extend(process_batch(batch))

    return results""",
                    "effort": "medium",
                    "impact": "high",
                },
                {
                    "title": "Clear Unused Objects",
                    "description": "Explicitly delete large objects when no longer needed.",
                    "code": """def handler(request):
    data = load_data()
    result = process(data)

    # Free memory
    del data

    return result""",
                    "effort": "low",
                    "impact": "medium",
                },
            ],
            "network": [
                {
                    "title": "Add Retry Logic",
                    "description": "Implement retry with exponential backoff for network calls.",
                    "code": """import time

def retry_with_backoff(func, max_retries=3):
    for i in range(max_retries):
        try:
            return func()
        except Exception as e:
            if i == max_retries - 1:
                raise
            time.sleep(2 ** i)

def handler(request):
    return retry_with_backoff(lambda: call_api(request))""",
                    "effort": "medium",
                    "impact": "high",
                },
                {
                    "title": "Add Connection Timeout",
                    "description": "Set explicit timeouts to avoid hanging on slow connections.",
                    "code": """import httpx

client = httpx.Client(timeout=5.0)

def handler(request):
    response = client.post(
        "https://api.example.com/endpoint",
        json=request
    )
    return response.json()""",
                    "effort": "low",
                    "impact": "medium",
                },
            ],
            "authentication": [
                {
                    "title": "Refresh Token Before Expiry",
                    "description": "Check token expiry and refresh before making API calls.",
                    "code": """import time

token = None
token_expiry = 0

def get_token():
    global token, token_expiry

    if time.time() > token_expiry - 300:  # Refresh 5 min early
        token, expires_at = refresh_token()
        token_expiry = expires_at

    return token

def handler(request):
    token = get_token()
    return call_api_with_token(token, request)""",
                    "effort": "medium",
                    "impact": "high",
                },
            ],
            "validation": [
                {
                    "title": "Add Input Validation",
                    "description": "Validate input at function entry with clear error messages.",
                    "code": """def validate_input(data):
    errors = []

    if "email" not in data:
        errors.append("email is required")
    elif not is_valid_email(data["email"]):
        errors.append("invalid email format")

    if errors:
        raise ValueError(", ".join(errors))

def handler(request):
    validate_input(request)
    # Continue with processing""",
                    "effort": "low",
                    "impact": "high",
                },
            ],
            "runtime": [
                {
                    "title": "Add Error Handling",
                    "description": "Wrap risky operations in try/catch blocks.",
                    "code": """def handler(request):
    try:
        result = risky_operation(request)
        return {"success": True, "data": result}
    except ValueError as e:
        return {"success": False, "error": str(e)}
    except Exception as e:
        return {"success": False, "error": "Internal error"}""",
                    "effort": "low",
                    "impact": "high",
                },
            ],
        }

    async def generate_suggestions(
        self,
        analysis: Dict[str, Any],
    ) -> List[Dict[str, Any]]:
        """Generate fix suggestions based on analysis.

        Args:
            analysis: Error analysis result

        Returns:
            List of fix suggestions with code examples
        """
        category = analysis.get("error_category", "unknown")
        root_cause = analysis.get("root_cause", "")

        # Get templates for this category
        templates = self._fix_templates.get(category, [])

        suggestions = []
        for template in templates:
            suggestion = {
                "id": f"suggestion-{category}-{template['title'].lower().replace(' ', '-')}",
                "title": template["title"],
                "description": template["description"],
                "code_example": template["code"],
                "effort": template["effort"],
                "impact": template["impact"],
                "applies_to": root_cause,
            }
            suggestions.append(suggestion)

        # If no templates matched, use LLM to generate suggestions
        if not suggestions:
            suggestions = await self._generate_llm_suggestions(analysis)

        return suggestions

    async def _generate_llm_suggestions(
        self,
        analysis: Dict[str, Any],
    ) -> List[Dict[str, Any]]:
        """Generate suggestions using LLM when no template matches.

        Args:
            analysis: Error analysis result

        Returns:
            List of suggestions
        """
        error_message = analysis.get("error_message", "")
        root_cause = analysis.get("root_cause", "")
        category = analysis.get("error_category", "unknown")

        prompt = f"""Based on this error analysis, provide specific fix suggestions:

Error Category: {category}
Error Message: {error_message}
Root Cause: {root_cause}

Provide 2-3 specific suggestions with:
1. Title (brief)
2. Description (what to do)
3. Code example (if applicable)

Format as JSON array of objects with keys: title, description, code_example"""

        try:
            provider_manager = get_provider_manager()
            provider = provider_manager.get_provider()

            response = await provider.complete(
                messages=[{"role": "user", "content": prompt}],
                temperature=0.3,
                max_tokens=500,
            )

            # Parse the response - this is a simplified approach
            import json
            try:
                suggestions = json.loads(response.content)
                return suggestions
            except json.JSONDecodeError:
                logger.warning("Failed to parse LLM response as JSON")
                return [{
                    "title": "Review Error Details",
                    "description": "Please review the error message and root cause analysis.",
                    "code_example": None,
                }]

        except Exception as e:
            logger.error(f"Failed to generate LLM suggestions: {e}")
            return [{
                "title": "Manual Investigation Required",
                "description": "Review the error details and consider consulting documentation.",
                "code_example": None,
            }]

    def get_documentation_links(
        self,
        category: str,
    ) -> List[Dict[str, str]]:
        """Get relevant documentation links for the error category.

        Args:
            category: Error category

        Returns:
            List of documentation links
        """
        docs_map = {
            "timeout": [
                {"title": "Function Timeout Configuration", "url": "/docs/functions/timeout"},
                {"title": "Performance Best Practices", "url": "/docs/performance"},
            ],
            "memory": [
                {"title": "Memory Management", "url": "/docs/functions/memory"},
                {"title": "Memory Profiling", "url": "/docs/debugging/memory"},
            ],
            "network": [
                {"title": "Network Configuration", "url": "/docs/functions/networking"},
                {"title": "Handling External APIs", "url": "/docs/integrations"},
            ],
            "authentication": [
                {"title": "Authentication Guide", "url": "/docs/auth"},
                {"title": "Managing Secrets", "url": "/docs/secrets"},
            ],
            "validation": [
                {"title": "Input Validation", "url": "/docs/functions/validation"},
                {"title": "Error Handling", "url": "/docs/functions/errors"},
            ],
            "runtime": [
                {"title": "Error Handling Guide", "url": "/docs/functions/errors"},
                {"title": "Debugging Functions", "url": "/docs/debugging"},
            ],
        }

        return docs_map.get(category, [])


_fix_suggester: Optional[FixSuggester] = None


def get_fix_suggester() -> FixSuggester:
    """Get the global fix suggester instance."""
    global _fix_suggester
    if _fix_suggester is None:
        _fix_suggester = FixSuggester()
    return _fix_suggester
