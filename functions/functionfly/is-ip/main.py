import socket


def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    version = event.get("version")

    if value is None:
        return {"ok": False, "error": "value is required"}

    val = str(value)
    is_v4 = False
    is_v6 = False

    try:
        socket.inet_pton(socket.AF_INET, val)
        is_v4 = True
    except (socket.error, OSError):
        pass

    try:
        socket.inet_pton(socket.AF_INET6, val)
        is_v6 = True
    except (socket.error, OSError):
        pass

    if version == 4:
        result = is_v4
    elif version == 6:
        result = is_v6
    else:
        result = is_v4 or is_v6

    detected_version = 4 if is_v4 else (6 if is_v6 else None)
    return {"ok": True, "value": value, "result": result, "version": detected_version}
