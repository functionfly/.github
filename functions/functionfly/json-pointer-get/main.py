def _pointer_get(doc, pointer):
    """RFC 6901 JSON Pointer"""
    if pointer == "":
        return doc
    if not pointer.startswith("/"):
        raise ValueError("JSON Pointer must be empty string or start with '/'")
    tokens = pointer[1:].split("/")
    current = doc
    for token in tokens:
        token = token.replace("~1", "/").replace("~0", "~")
        if isinstance(current, dict):
            if token not in current:
                raise KeyError(f"key not found: {token!r}")
            current = current[token]
        elif isinstance(current, list):
            try:
                idx = int(token)
            except ValueError:
                raise ValueError(f"list index must be integer, got: {token!r}")
            if not 0 <= idx < len(current):
                raise IndexError(f"index {idx} out of range (len={len(current)})")
            current = current[idx]
        else:
            raise TypeError(f"cannot traverse into {type(current).__name__}")
    return current


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    pointer = event.get("pointer", "")
    if data is None:
        return {"ok": False, "error": "data is required"}
    try:
        value = _pointer_get(data, str(pointer))
        return {"ok": True, "result": value}
    except (KeyError, IndexError, ValueError, TypeError) as e:
        return {"ok": False, "error": str(e)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
