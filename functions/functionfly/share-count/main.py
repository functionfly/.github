def handler(event):
    url = event.get("url") if isinstance(event, dict) else None
    platform = event.get("platform", "all")
    if not url:
        return {"ok": False, "error": "url is required"}
    try:
        # Note: Real share counts require platform API keys and auth.
        # This function returns the structure for integrating with social APIs.
        platforms = ["facebook", "twitter", "linkedin", "pinterest", "reddit"]
        if platform != "all" and platform not in platforms:
            return {"ok": False, "error": f"platform must be one of {platforms} or 'all'"}
        note = "Real share counts require platform API authentication. This returns the data schema for integration."
        shares = {p: None for p in (platforms if platform == "all" else [platform])}
        return {
            "ok": True,
            "result": shares,
            "url": str(url),
            "shares": shares,
            "note": note,
            "total": None
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
