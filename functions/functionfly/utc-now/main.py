from datetime import datetime, timezone

def handler(event):
    now = datetime.now(timezone.utc)
    return {"ok": True, "iso": now.isoformat(), "timestamp": int(now.timestamp())}
