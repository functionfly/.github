import socket


def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    if value is None:
        return {"ok": False, "error": "value is required"}
    try:
        socket.inet_pton(socket.AF_INET6, str(value))
        result = True
    except (socket.error, OSError):
        result = False
    return {"ok": True, "value": value, "result": result}
