import re


def handler(event):
    html = event.get("html") if isinstance(event, dict) else None
    base_url = event.get("base_url", "")
    if not html:
        return {"ok": False, "error": "html is required"}
    try:
        text = str(html)
        favicons = []
        # Look for link rel icons
        for m in re.finditer(r'<link\s([^>]+)>', text, re.I):
            attrs = m.group(1)
            rel_m = re.search(r'rel=["\']([^"\']*icon[^"\']*)["\']', attrs, re.I)
            if rel_m:
                href_m = re.search(r'href=["\']([^"\']+)["\']', attrs, re.I)
                size_m = re.search(r'sizes=["\']([^"\']+)["\']', attrs, re.I)
                if href_m:
                    favicons.append({
                        "url": href_m.group(1),
                        "rel": rel_m.group(1),
                        "sizes": size_m.group(1) if size_m else None
                    })
        # Default favicon
        if base_url:
            from urllib.parse import urlparse
            parsed = urlparse(str(base_url))
            default = f"{parsed.scheme}://{parsed.netloc}/favicon.ico"
        else:
            default = "/favicon.ico"
        primary = favicons[0]["url"] if favicons else default
        return {
            "ok": True,
            "result": primary,
            "favicon_url": primary,
            "all_favicons": favicons,
            "count": len(favicons),
            "default_favicon": default
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
