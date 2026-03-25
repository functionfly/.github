import re
from urllib.parse import urlparse


def handler(event):
    url = event.get("url") if isinstance(event, dict) else None
    if not url:
        return {"ok": False, "error": "url is required"}
    try:
        parsed = urlparse(str(url))
        domain = parsed.netloc.lstrip("www.")
        path_parts = [p for p in parsed.path.strip('/').split('/') if p]
        slug = path_parts[-1] if path_parts else ""
        clean_slug = re.sub(r'[-_]', ' ', re.sub(r'\.[^.]+$', '', slug))
        inferred_title = clean_slug.title() if clean_slug else domain
        return {
            "ok": True,
            "result": {"url": str(url), "domain": domain},
            "url": str(url),
            "domain": domain,
            "scheme": parsed.scheme,
            "path": parsed.path,
            "query": parsed.query,
            "fragment": parsed.fragment,
            "slug": slug,
            "inferred_title": inferred_title,
            "note": "Actual title/og data requires HTTP fetch — use og-tags-extract with fetched HTML"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
