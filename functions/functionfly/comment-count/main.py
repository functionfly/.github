def handler(event):
    content_id = event.get("content_id") if isinstance(event, dict) else None
    platform = event.get("platform", "all")
    if not content_id:
        return {"ok": False, "error": "content_id is required"}
    try:
        note = "Real comment counts require platform API authentication. This returns the data schema for integration."
        platforms = ["facebook", "instagram", "youtube", "reddit", "disqus"]
        comments = {p: None for p in (platforms if platform == "all" else [platform])}
        return {
            "ok": True,
            "result": comments,
            "content_id": str(content_id),
            "platform": platform,
            "comments": comments,
            "note": note,
            "total": None
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
