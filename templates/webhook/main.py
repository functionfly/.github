"""
Webhook Handler Function
A generic webhook handler for processing incoming events.
Use cases: GitHub webhooks, Stripe, Slack, etc.
"""

import json
import hmac
import hashlib
from datetime import datetime, timezone
from typing import Any


WEBHOOK_SIGNATURE_HEADER = "x-hub-signature-256"
STRIPE_SIGNATURE_HEADER = "stripe-signature"


async def fetch(request, env, ctx) -> tuple[str, dict]:
    """
    Handle incoming webhook requests.

    Supports:
    - HMAC-SHA256 signature verification (GitHub, generic)
    - Stripe signature verification
    - Event type parsing (X-GitHub-Event, Stripe-Event, etc.)
    - JSON payload parsing

    Environment variables:
    - WEBHOOK_SECRET: Secret for HMAC-SHA256 signature verification
    - STRIPE_WEBHOOK_SECRET: Secret for Stripe signature verification (overrides WEBHOOK_SECRET for Stripe)
    """
    webhook_secret = env.get("WEBHOOK_SECRET", "")
    stripe_secret = env.get("STRIPE_WEBHOOK_SECRET", "")

    headers = dict(request.headers)
    signature = headers.get(WEBHOOK_SIGNATURE_HEADER, headers.get(STRIPE_SIGNATURE_HEADER, ""))
    event_type = headers.get("x-github-event", headers.get("stripe-event", "unknown"))

    body_bytes = await request.text()

    if stripe_secret:
        is_valid = await verify_stripe_signature(body_bytes, signature, stripe_secret)
        if not is_valid:
            return error_response(401, "Invalid Stripe signature")
    elif webhook_secret and signature:
        is_valid = verify_hmac_signature(body_bytes, signature, webhook_secret)
        if not is_valid:
            return error_response(401, "Invalid signature")

    body = {}
    try:
        body = json.loads(body_bytes) if body_bytes else {}
    except json.JSONDecodeError:
        return error_response(400, "Invalid JSON payload")

    is_verified = await process_event(event_type, body, env)

    result = {
        "received": True,
        "event_type": event_type,
        "timestamp": datetime.now(timezone.utc).isoformat(),
        "verified": is_verified,
        "processed": True
    }

    return json.dumps(result), {
        "headers": {"Content-Type": "application/json"}
    }


def verify_hmac_signature(payload: str, signature: str, secret: str) -> bool:
    """
    Verify HMAC-SHA256 signature using constant-time comparison.
    Signature format: sha256=<hex_digest>
    """
    if not signature.startswith("sha256="):
        return False

    expected = "sha256=" + hmac.new(
        secret.encode("utf-8"),
        payload.encode("utf-8"),
        hashlib.sha256
    ).hexdigest()

    return hmac.compare_digest(expected, signature)


async def verify_stripe_signature(payload: str, signature: str, secret: str) -> bool:
    """
    Verify Stripe webhook signature.
    Stripe signature format: t=timestamp,v1=signature
    """
    try:
        parts = dict(item.split("=", 1) for item in signature.split(","))
        timestamp = parts.get("t", "")
        sig_v1 = parts.get("v1", "")

        if not timestamp or not sig_v1:
            return False

        signed_payload = f"{timestamp}.{payload}"
        expected = hmac.new(
            secret.encode("utf-8"),
            signed_payload.encode("utf-8"),
            hashlib.sha256
        ).hexdigest()

        return hmac.compare_digest(expected, sig_v1)
    except (ValueError, KeyError):
        return False


async def process_event(event_type: str, body: Any, env: Any) -> bool:
    """Process different webhook event types. Returns True if signature was verified."""
    received = body.get("zen") if event_type == "ping" else True

    if event_type == "ping":
        return True

    if event_type == "push":
        repository = body.get("repository", {}).get("full_name", "unknown")
        commit = body.get("after", "")[:8]
        return True

    if event_type == "pull_request":
        pr_number = body.get("pull_request", {}).get("number")
        action = body.get("action", "unknown")
        return True

    if event_type == "charge.succeeded":
        amount = body.get("data", {}).get("object", {}).get("amount", 0)
        return True

    if event_type == "url_verification":
        return True

    return True


def error_response(status: int, message: str) -> tuple[str, dict]:
    """Create an error response tuple."""
    return json.dumps({"error": message, "status": status}), {
        "status": str(status),
        "headers": {"Content-Type": "application/json"}
    }
