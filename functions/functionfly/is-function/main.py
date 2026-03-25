def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    if value is None and "value" not in event:
        return {"ok": False, "error": "value is required"}
    # In JSON context, functions cannot be serialized; this checks if a string
    # looks like a function definition
    import re
    if isinstance(value, str):
        func_re = re.compile(
            r'^(function\s*\w*\s*\(|const\s+\w+\s*=\s*(\(.*\)|[^=]+)\s*=>|\w+\s*=\s*function|\(?.*\)?\s*=>)',
            re.MULTILINE
        )
        result = bool(func_re.match(value.strip()))
    else:
        result = callable(value)
    return {"ok": True, "value": value if not callable(value) else str(value), "result": result}
