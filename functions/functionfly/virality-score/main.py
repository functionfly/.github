import math


def handler(event):
    shares = event.get("shares", 0) if isinstance(event, dict) else 0
    likes = event.get("likes", 0)
    comments = event.get("comments", 0)
    views = event.get("views")
    time_hours = event.get("time_hours", 24)
    try:
        s, l, c = float(shares), float(likes), float(comments)
        h = max(float(time_hours), 0.1)
        # Shares weighted 3x because they amplify reach, comments 1.5x for high engagement signal
        weighted = (s * 3) + (l * 1) + (c * 1.5)
        # Velocity: engagement per hour
        velocity = weighted / h
        if views and float(views) > 0:
            v = float(views)
            viral_ratio = (s / v) * 100
        else:
            viral_ratio = None
        # Normalize to 0-100 score using log scale
        score = round(min(100, math.log1p(velocity) * 10), 2)
        if score < 20:
            category = "low"
        elif score < 40:
            category = "moderate"
        elif score < 60:
            category = "high"
        elif score < 80:
            category = "very_high"
        else:
            category = "viral"
        return {
            "ok": True,
            "result": score,
            "virality_score": score,
            "category": category,
            "velocity": round(velocity, 4),
            "viral_ratio_percent": round(viral_ratio, 4) if viral_ratio is not None else None,
            "shares": s,
            "time_hours": h
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
