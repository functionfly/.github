import time


def handler(event):
    cart_created_at = event.get("cart_created_at") if isinstance(event, dict) else None
    last_activity_at = event.get("last_activity_at")
    abandonment_threshold_minutes = int(event.get("abandonment_threshold_minutes", 30))
    recovery_window_hours = int(event.get("recovery_window_hours", 24))
    if cart_created_at is None:
        return {"ok": False, "error": "cart_created_at is required (unix timestamp)"}
    try:
        now = int(time.time())
        created = int(cart_created_at)
        last_activity = int(last_activity_at) if last_activity_at else created
        idle_seconds = now - last_activity
        idle_minutes = idle_seconds // 60
        is_abandoned = idle_minutes >= abandonment_threshold_minutes
        age_hours = (now - created) / 3600
        in_recovery_window = age_hours <= recovery_window_hours
        minutes_until_abandoned = max(0, abandonment_threshold_minutes - idle_minutes)
        return {
            "ok": True,
            "result": is_abandoned,
            "is_abandoned": is_abandoned,
            "idle_minutes": idle_minutes,
            "age_hours": round(age_hours, 2),
            "in_recovery_window": in_recovery_window,
            "minutes_until_abandoned": minutes_until_abandoned,
            "should_send_recovery_email": is_abandoned and in_recovery_window
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
