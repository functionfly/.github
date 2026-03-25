import re


def handler(event):
    html = event.get("html") if isinstance(event, dict) else None
    if not html:
        return {"ok": False, "error": "html is required"}
    try:
        text = str(html)
        tags = {}
        for m in re.finditer(r'<meta\s+([^>]+?)/?>', text, re.I | re.S):
            attrs_str = m.group(1)
            name_m = re.search(r'(?:name|property|http-equiv)=["\']([^"\']+)["\']', attrs_str, re.I)
            content_m = re.search(r'content=["\']([^"\']*)["\']', attrs_str, re.I)
            if name_m and content_m:
                tags[name_m.group(1)] = content_m.group(1)
        title_m = re.search(r'<title[^>]*>([^<]+)</title>', text, re.I)
        title = title_m.group(1).strip() if title_m else None
        return {"ok": True, "result": tags, "meta_tags": tags, "title": title, "count": len(tags)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
