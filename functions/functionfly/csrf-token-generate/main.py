import base64
import hashlib
import hmac
import secrets


def handler(event):
    if isinstance(event, dict):
        secret = event.get("secret", "")
        nbytes = event.get("bytes", 32)
    else:
        secret, nbytes = "", 32

    try:
        nbytes = max(16, min(64, int(nbytes)))
    except (TypeError, ValueError):
        nbytes = 32

    raw = secrets.token_bytes(nbytes)
    token_b64 = base64.urlsafe_b64encode(raw).decode("ascii").rstrip("=")
    if secret:
        sig = hmac.new(secret.encode("utf-8") if isinstance(secret, str) else secret, raw, hashlib.sha256).hexdigest()
        token_b64 = token_b64 + "." + sig
    return {"ok": True, "token": token_b64}
