import urllib.parse


def handler(event):
    url = event.get("url") if isinstance(event, dict) else None
    title = event.get("title", "")
    summary = event.get("summary", "")
    source = event.get("source", "")
    if not url:
        return {"ok": False, "error": "url is required"}
    try:
        params = {"mini": "true", "url": str(url)}
        if title: params["title"] = str(title)
        if summary: params["summary"] = str(summary)
        if source: params["source"] = str(source)
        share_url = "https://www.linkedin.com/shareArticle?" + urllib.parse.urlencode(params)
        html = f'<a href="{share_url}" target="_blank" rel="noopener noreferrer">Share on LinkedIn</a>'
        return {"ok": True, "result": share_url, "share_url": share_url, "html": html}
    except Exception as e:
        return {"ok": False, "error": str(e)}
