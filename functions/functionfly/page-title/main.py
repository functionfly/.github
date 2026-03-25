import re


def handler(event):
    html = event.get("html") if isinstance(event, dict) else None
    if not html:
        return {"ok": False, "error": "html is required"}
    try:
        text = str(html)
        title = None
        # 1. <title> tag
        m = re.search(r'<title[^>]*>(.*?)</title>', text, re.I | re.S)
        if m:
            title = re.sub(r'\s+', ' ', m.group(1).strip())
        # 2. og:title
        og_title = None
        for pattern in [
            r'<meta\s+(?:[^>]*?\s)?property=["\']og:title["\']\s+content=["\'](.*?)["\']',
            r'<meta\s+(?:[^>]*?\s)?content=["\'](.*?)["\'](?:[^>]*?\s)?property=["\']og:title["\']',
        ]:
            m2 = re.search(pattern, text, re.I | re.S)
            if m2:
                og_title = m2.group(1).strip()
                break
        return {
            "ok": True,
            "result": title,
            "title": title,
            "og_title": og_title,
            "best_title": og_title or title
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
