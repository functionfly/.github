import re
import os


def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    allow_absolute = event.get("allow_absolute", False)
    separator = event.get("separator", "/")

    if not value:
        return {"ok": False, "error": "value is required"}
    if not isinstance(value, str):
        return {"ok": False, "error": "value must be a string"}

    path = value
    # Normalize separators
    path = path.replace("\\", "/")
    # Remove null bytes and control characters
    path = re.sub(r'[\x00-\x1f]', '', path)
    # Resolve path traversal (..)
    parts = path.split("/")
    resolved = []
    for part in parts:
        if part == "..":
            if resolved and resolved[-1] != "":
                resolved.pop()
        elif part == ".":
            pass
        else:
            resolved.append(part)
    path = "/".join(resolved)
    # Remove leading slash if not allowed
    if not allow_absolute:
        path = path.lstrip("/")
    # Convert separator
    if separator != "/":
        path = path.replace("/", separator)
    # Remove trailing separator
    path = path.rstrip(separator)

    if not path:
        path = "."

    return {"ok": True, "value": value, "result": path, "changed": path != value}
