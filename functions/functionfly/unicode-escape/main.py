import re


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    non_ascii_only = event.get("non_ascii_only", True)

    if data is None:
        return {"ok": False, "error": "data is required"}
    try:
        s = str(data)
        if non_ascii_only:
            result = "".join(
                f"\\u{ord(c):04x}" if ord(c) > 127 else c
                for c in s
            )
        else:
            result = "".join(f"\\u{ord(c):04x}" for c in s)
        return {"ok": True, "result": result}
    except Exception as e:
        return {"ok": False, "error": str(e)}
