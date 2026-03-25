import re


def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    pattern = event.get("pattern")
    flags_str = event.get("flags", "")
    full_match = event.get("full_match", False)

    if value is None:
        return {"ok": False, "error": "value is required"}
    if not pattern:
        return {"ok": False, "error": "pattern is required"}

    flag_map = {"i": re.IGNORECASE, "m": re.MULTILINE, "s": re.DOTALL, "x": re.VERBOSE}
    flags = 0
    for f in str(flags_str):
        flags |= flag_map.get(f, 0)

    try:
        compiled = re.compile(str(pattern), flags)
        if full_match:
            m = compiled.fullmatch(str(value))
        else:
            m = compiled.search(str(value))
        result = bool(m)
        groups = list(m.groups()) if m and m.groups() else []
        match_str = m.group(0) if m else None
    except re.error as e:
        return {"ok": False, "error": f"Invalid regex pattern: {str(e)}"}

    return {"ok": True, "value": value, "result": result, "match": match_str, "groups": groups}
