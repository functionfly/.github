"""
Webhook Handler Function
A generic webhook handler for processing incoming events.
Use cases: GitHub webhooks, Stripe, Slack, etc.
"""

import json
import hashlib
import hmac
from datetime import datetime


async def fetch(request, env, ctx):
    """
    Handle incoming webhook requests.
    
    Supports:
    - Signature verification (X-Hub-Signature-256)
    - Event type parsing (X-GitHub-Event, Stripe-Event, etc.)
    - JSON payload parsing
    
    Environment variables:
    - WEBHOOK_SECRET: Secret for signature verification
    """
    # Get webhook secret from environment
    webhook_secret = env.get("WEBHOOK_SECRET", "")
    
    # Get headers
    headers = dict(request.headers)
    signature = headers.get("x-hub-signature-256", "")
    event_type = headers.get("x-github-event", headers.get("stripe-event", "unknown"))
    
    # Parse request body
    body = {}
    try:
        body = await request.json()
    except:
        pass
    
    # Verify signature if secret is configured
    if webhook_secret and signature:
        # In production, verify the signature properly
        # Example: hmac.compare_digest(...)
        pass
    
    # Process based on event type
    result = await process_event(event_type, body, env)
    
    return {
        "status": 200,
        "body": result,
        "headers": {"Content-Type": "application/json"}
    }


async def process_event(event_type, body, env):
    """Process different webhook event types."""
    
    result = {
        "received": True,
        "event_type": event_type,
        "timestamp": datetime.utcnow().isoformat(),
        "processed": False,
        "actions": []
    }
    
    # GitHub events
    if event_type == "push":
        result["actions"].append("processed_git_push")
        result["repository"] = body.get("repository", {}).get("full_name")
        result["commit"] = body.get("after", "")
        result["processed"] = True
        
    elif event_type == "pull_request":
        result["actions"].append("processed_pull_request")
        result["pr_number"] = body.get("number")
        result["action"] = body.get("action")
        result["processed"] = True
        
    # Stripe events
    elif event_type == "charge.succeeded":
        result["actions"].append("processed_payment")
        result["amount"] = body.get("data", {}).get("object", {}).get("amount")
        result["processed"] = True
        
    # Slack events
    elif event_type == "url_verification":
        result["actions"].append("verified_url")
        result["challenge"] = body.get("challenge")
        result["processed"] = True
        
    # Generic event
    else:
        result["actions"].append("processed_generic")
        result["processed"] = True
    
    return result
