import re


def handler(event):
    html = event.get("html") if isinstance(event, dict) else None
    if not html:
        return {"ok": False, "error": "html is required"}
    try:
        text = str(html)
        og = {}
        for m in re.finditer(r'<meta\s+(?:[^>]*?\s)?property=["\']og:([^"\']+)["\'][^>]*?content=["\']([^"\']*)["\']', text, re.I):
            og[m.group(1)] = m.group(2)
        for m in re.finditer(r'<meta\s+(?:[^>]*?\s)?content=["\']([^"\']*)["\'][^>]*?property=["\']og:([^"\']+)["\']', text, re.I):
            og[m.group(2)] = m.group(1)
        return {"ok": True, "result": og, "og_tags": og, "count": len(og), "has_og": len(og) > 0}
    except Exception as e:
        return {"ok": False, "error": str(e)}
