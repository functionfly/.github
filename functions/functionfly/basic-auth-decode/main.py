import base64


def handler(event):
    credentials = event.get("credentials") if isinstance(event, dict) else None

    if not credentials:
        return {"ok": False, "error": "credentials is required"}
    if not isinstance(credentials, str):
        return {"ok": False, "error": "credentials must be a string"}

    # Strip "Basic " prefix if present
    creds = credentials.strip()
    if creds.lower().startswith("basic "):
        creds = creds[6:].strip()

    try:
        decoded = base64.b64decode(creds).decode("utf-8")
    except Exception as e:
        return {"ok": False, "error": f"Failed to decode Base64: {str(e)}"}

    if ":" not in decoded:
        return {"ok": False, "error": "Decoded credentials do not contain a colon separator"}

    username, password = decoded.split(":", 1)
    return {"ok": True, "username": username, "password": password}
