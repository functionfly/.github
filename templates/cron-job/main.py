"""
Cron Job Function
A scheduled function that runs periodically.
Use cases: Data cleanup, report generation, cache refresh, etc.
"""

import json
from datetime import datetime


async def fetch(request, env, ctx):
    """
    Handle scheduled execution.
    
    This function runs automatically based on the schedule defined
    in functionfly.jsonc. It receives a trigger event with timing info.
    
    Schedule presets:
        "*/5 * * * *"    - Every 5 minutes
        "0 * * * *"      - Every hour
        "0 0 * * *"      - Every day at midnight
        "0 0 * * 1-5"    - Weekdays at midnight
    """
    # Check if this is a scheduled execution
    # Scheduled functions receive a special trigger event
    body = {}
    try:
        body = await request.json()
    except:
        pass
    
    trigger = body.get("trigger", "manual")
    timestamp = body.get("timestamp", datetime.utcnow().isoformat())
    
    # Example: Perform scheduled task
    # In production, replace with actual task logic:
    # - Database cleanup
    # - Generate reports
    # - Refresh caches
    # - Send notifications
    
    result = {
        "status": "success",
        "trigger": trigger,
        "executed_at": timestamp,
        "message": "Scheduled job completed successfully",
        "tasks": {
            "cleanup": {"status": "completed", "items_processed": 0},
            "report": {"status": "completed", "generated_at": datetime.utcnow().isoformat()},
            "cache": {"status": "completed", "entries_refreshed": 0}
        }
    }
    
    return {
        "status": 200,
        "body": result,
        "headers": {
            "Content-Type": "application/json",
            "X-FunctionFly-Template": "cron-job"
        }
    }


# Example: Manual trigger handler
async def handle_manual(request, env, ctx):
    """Handle manual invocation for testing."""
    return await fetch(request, env, ctx)
