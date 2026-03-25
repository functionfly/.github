import re


def handler(event):
    html = event.get("html") if isinstance(event, dict) else None
    if not html:
        return {"ok": False, "error": "html is required"}
    try:
        text = str(html)
        tc = {}
        for m in re.finditer(r'<meta\s+(?:[^>]*?\s)?name=["\']twitter:([^"\']+)["\'][^>]*?content=["\']([^"\']*)["\']', text, re.I):
            tc[m.group(1)] = m.group(2)
        for m in re.finditer(r'<meta\s+(?:[^>]*?\s)?content=["\']([^"\']*)["\'][^>]*?name=["\']twitter:([^"\']+)["\']', text, re.I):
            tc[m.group(2)] = m.group(1)
        return {"ok": True, "result": tc, "twitter_card": tc, "card_type": tc.get("card"), "has_card": len(tc) > 0}
    except Exception as e:
        return {"ok": False, "error": str(e)}
