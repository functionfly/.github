def handler(event):
    content_id = event.get("content_id") if isinstance(event, dict) else None
    platform = event.get("platform", "all")
    if not content_id:
        return {"ok": False, "error": "content_id is required"}
    try:
        note = "Real like counts require platform API authentication. This returns the data schema for integration."
        platforms = ["facebook", "twitter", "instagram", "tiktok", "youtube"]
        likes = {p: None for p in (platforms if platform == "all" else [platform])}
        return {
            "ok": True,
            "result": likes,
            "content_id": str(content_id),
            "platform": platform,
            "likes": likes,
            "note": note,
            "total": None
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
