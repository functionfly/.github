import re


def handler(event):
    html = event.get("html") if isinstance(event, dict) else None
    if not html:
        return {"ok": False, "error": "html is required"}
    try:
        text = str(html)
        meta_desc = None
        for pattern in [
            r'<meta\s+(?:[^>]*?\s)?name=["\']description["\']\s+content=["\'](.*?)["\']',
            r'<meta\s+(?:[^>]*?\s)?content=["\'](.*?)["\'](?:[^>]*?\s)?name=["\']description["\']',
        ]:
            m = re.search(pattern, text, re.I | re.S)
            if m:
                meta_desc = m.group(1).strip()
                break
        og_desc = None
        for pattern in [
            r'<meta\s+(?:[^>]*?\s)?property=["\']og:description["\']\s+content=["\'](.*?)["\']',
            r'<meta\s+(?:[^>]*?\s)?content=["\'](.*?)["\'](?:[^>]*?\s)?property=["\']og:description["\']',
        ]:
            m = re.search(pattern, text, re.I | re.S)
            if m:
                og_desc = m.group(1).strip()
                break
        description = meta_desc or og_desc
        return {
            "ok": True,
            "result": description,
            "description": description,
            "meta_description": meta_desc,
            "og_description": og_desc,
            "length": len(description) if description else 0
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
