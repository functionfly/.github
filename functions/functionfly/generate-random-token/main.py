import secrets

def handler(event):
    nbytes = event.get("bytes", 32) if isinstance(event, dict) else 32
    try:
        nbytes = max(8, min(512, int(nbytes)))
    except (TypeError, ValueError):
        nbytes = 32
    return {"ok": True, "token": secrets.token_urlsafe(nbytes)}
