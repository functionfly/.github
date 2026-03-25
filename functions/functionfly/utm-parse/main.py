def handler(event):
    url = event.get("url") if isinstance(event, dict) else None
    if not url:
        return {"ok": False, "error": "url is required"}
    try:
        from urllib.parse import urlparse, parse_qs
        parsed = urlparse(str(url))
        qs = parse_qs(parsed.query)
        utm_keys = ["utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content", "utm_id"]
        utm_params = {k: qs[k][0] for k in utm_keys if k in qs}
        has_utm = len(utm_params) > 0
        return {
            "ok": True,
            "result": utm_params,
            "utm_params": utm_params,
            "has_utm": has_utm,
            "base_url": f"{parsed.scheme}://{parsed.netloc}{parsed.path}",
            "source": utm_params.get("utm_source"),
            "medium": utm_params.get("utm_medium"),
            "campaign": utm_params.get("utm_campaign"),
            "term": utm_params.get("utm_term"),
            "content": utm_params.get("utm_content")
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
