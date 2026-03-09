import re


def handler(event):
    if isinstance(event, dict):
        text = event.get("text", event.get("data", ""))
    else:
        text = str(event) if event is not None else ""

    if text is None:
        return {"ok": False, "error": "Input 'text' is required"}

    # Match common phone patterns: (xxx) xxx-xxxx, xxx-xxx-xxxx, xxx.xxx.xxxx, +1 xxx..., digits with optional separators
    pattern = r"\+?[\d\s\-\.\(\)]{10,20}"
    candidates = re.findall(pattern, str(text))
    seen = set()
    phones = []
    for p in candidates:
        digits = re.sub(r"\D", "", p)
        if len(digits) >= 10 and len(digits) <= 15 and p.strip() not in seen:
            seen.add(p.strip())
            phones.append(p.strip())
    return {"ok": True, "phones": phones, "count": len(phones)}

