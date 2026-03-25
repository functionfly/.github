import re


def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    if value is None:
        return {"ok": False, "error": "value is required"}

    if isinstance(value, str):
        # Check for /pattern/flags notation
        js_re = re.match(r'^/(.+)/([gimsuy]*)$', value)
        if js_re:
            pattern = js_re.group(1)
            flags_str = js_re.group(2)
        else:
            pattern = value
            flags_str = ""
        try:
            re.compile(pattern)
            result = True
        except re.error:
            result = False
        return {"ok": True, "value": value, "result": result, "pattern": pattern if result else None, "flags": flags_str}

    return {"ok": True, "value": str(value), "result": False}
