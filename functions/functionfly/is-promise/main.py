import re

PROMISE_RE = re.compile(
    r'(new\s+Promise\s*\(|Promise\.(resolve|reject|all|race|allSettled|any)\s*\(|\.then\s*\(|async\s+function|async\s*\()',
    re.IGNORECASE
)


def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    if value is None:
        return {"ok": False, "error": "value is required"}
    if not isinstance(value, str):
        return {"ok": True, "value": value, "result": False}
    result = bool(PROMISE_RE.search(str(value).strip()))
    return {"ok": True, "value": value, "result": result}
