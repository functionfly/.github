def handler(event):
    likes = event.get("likes", 0) if isinstance(event, dict) else 0
    comments = event.get("comments", 0)
    shares = event.get("shares", 0)
    saves = event.get("saves", 0)
    followers = event.get("followers")
    impressions = event.get("impressions")
    metric_base = event.get("metric_base", "followers")
    try:
        l, c, s, sv = float(likes), float(comments), float(shares), float(saves)
        total_engagements = l + c + s + sv
        if metric_base == "impressions":
            if not impressions or float(impressions) <= 0:
                return {"ok": False, "error": "impressions required and must be > 0 when metric_base is impressions"}
            base = float(impressions)
        else:
            if not followers or float(followers) <= 0:
                return {"ok": False, "error": "followers required and must be > 0 when metric_base is followers"}
            base = float(followers)
        rate = round((total_engagements / base) * 100, 4)
        if rate < 1:
            quality = "low"
        elif rate < 3.5:
            quality = "average"
        elif rate < 6:
            quality = "good"
        else:
            quality = "excellent"
        return {
            "ok": True,
            "result": rate,
            "engagement_rate": rate,
            "engagement_rate_percent": f"{rate}%",
            "total_engagements": total_engagements,
            "quality": quality,
            "base": base,
            "metric_base": metric_base
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
