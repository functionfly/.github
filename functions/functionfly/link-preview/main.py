import re


def _extract_meta(html, prop_attr, prop_val):
    for pattern in [
        rf'<meta\s+(?:[^>]*?\s)?{prop_attr}=["\']{prop_val}["\']\s+content=["\'](.*?)["\']',
        rf'<meta\s+(?:[^>]*?\s)?content=["\'](.*?)["\'](?:[^>]*?\s)?{prop_attr}=["\']{prop_val}["\']',
    ]:
        m = re.search(pattern, html, re.I | re.S)
        if m:
            return m.group(1).strip()
    return None


def handler(event):
    html = event.get("html") if isinstance(event, dict) else None
    url = event.get("url", "")
    if not html:
        return {"ok": False, "error": "html is required"}
    try:
        text = str(html)
        # Title
        title = _extract_meta(text, "property", "og:title")
        if not title:
            m = re.search(r'<title[^>]*>(.*?)</title>', text, re.I | re.S)
            title = re.sub(r'\s+', ' ', m.group(1).strip()) if m else None
        # Description
        description = _extract_meta(text, "property", "og:description") or _extract_meta(text, "name", "description")
        # Image
        image = _extract_meta(text, "property", "og:image") or _extract_meta(text, "name", "twitter:image")
        # Site name
        site_name = _extract_meta(text, "property", "og:site_name")
        # Canonical URL
        can_m = re.search(r'<link\s[^>]*rel=["\']canonical["\'][^>]*href=["\'](.*?)["\']', text, re.I)
        canonical = can_m.group(1) if can_m else url
        return {
            "ok": True,
            "result": {"title": title, "description": description, "image": image, "url": canonical},
            "title": title,
            "description": description,
            "image": image,
            "url": canonical,
            "site_name": site_name,
            "has_preview": bool(title and (description or image))
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
