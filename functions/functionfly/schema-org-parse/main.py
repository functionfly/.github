import re, json


def handler(event):
    html = event.get("html") if isinstance(event, dict) else None
    data = event.get("data")
    if not html and not data:
        return {"ok": False, "error": "html or data is required"}
    try:
        if data:
            schema = data if isinstance(data, (dict, list)) else json.loads(str(data))
        else:
            text = str(html)
            matches = re.findall(r'<script[^>]+type=["\']application/ld\+json["\'][^>]*>(.*?)</script>', text, re.I | re.S)
            if not matches:
                return {"ok": True, "result": None, "schemas": [], "count": 0, "note": "No JSON-LD found"}
            schema = json.loads(matches[0].strip())
        schemas = schema if isinstance(schema, list) else [schema]
        parsed = []
        for s in schemas:
            schema_type = s.get("@type", "Unknown")
            parsed.append({"@type": schema_type, "context": s.get("@context", "https://schema.org"), "data": s})
        return {"ok": True, "result": parsed, "schemas": parsed, "count": len(parsed), "types": [p["@type"] for p in parsed]}
    except Exception as e:
        return {"ok": False, "error": str(e)}
