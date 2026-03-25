try:
    from urllib.parse import urlparse, urlencode, urlunparse, parse_qs
except ImportError:
    pass


def handler(event):
    url = event.get("url") if isinstance(event, dict) else None
    utm_source = event.get("utm_source")
    utm_medium = event.get("utm_medium")
    utm_campaign = event.get("utm_campaign")
    utm_term = event.get("utm_term")
    utm_content = event.get("utm_content")
    if not url:
        return {"ok": False, "error": "url is required"}
    if not utm_source:
        return {"ok": False, "error": "utm_source is required"}
    if not utm_medium:
        return {"ok": False, "error": "utm_medium is required"}
    if not utm_campaign:
        return {"ok": False, "error": "utm_campaign is required"}
    try:
        from urllib.parse import urlparse, urlencode, urlunparse, parse_qs
        parsed = urlparse(str(url))
        params = {}
        params["utm_source"] = str(utm_source)
        params["utm_medium"] = str(utm_medium)
        params["utm_campaign"] = str(utm_campaign)
        if utm_term:
            params["utm_term"] = str(utm_term)
        if utm_content:
            params["utm_content"] = str(utm_content)
        existing = parse_qs(parsed.query)
        combined = {k: v[0] for k, v in existing.items()}
        combined.update(params)
        new_query = urlencode(combined)
        result_url = urlunparse((parsed.scheme, parsed.netloc, parsed.path, parsed.params, new_query, parsed.fragment))
        return {
            "ok": True,
            "result": result_url,
            "url": result_url,
            "utm_params": params
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
