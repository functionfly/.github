import re, json


def handler(event):
    html = event.get("html") if isinstance(event, dict) else None
    if not html:
        return {"ok": False, "error": "html is required"}
    try:
        text = str(html)
        matches = re.findall(r'<script[^>]+type=["\']application/ld\+json["\'][^>]*>(.*?)</script>', text, re.I | re.S)
        items = []
        for m in matches:
            try:
                items.append(json.loads(m.strip()))
            except json.JSONDecodeError:
                items.append({"_parse_error": True, "_raw": m.strip()[:200]})
        return {"ok": True, "result": items, "json_ld": items, "count": len(items)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
