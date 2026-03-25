import base64


def handler(event):
    username = event.get("username") if isinstance(event, dict) else None
    password = event.get("password", "")

    if username is None:
        return {"ok": False, "error": "username is required"}
    if not isinstance(username, str):
        return {"ok": False, "error": "username must be a string"}
    if not isinstance(password, str):
        return {"ok": False, "error": "password must be a string"}

    raw = f"{username}:{password}"
    encoded = base64.b64encode(raw.encode("utf-8")).decode("utf-8")
    header = f"Basic {encoded}"

    return {"ok": True, "credentials": encoded, "header": header}
