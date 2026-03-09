import base64
import hashlib
import hmac


def handler(event):
    if isinstance(event, dict):
        token = event.get("token", "")
        expected = event.get("expected", "")
        secret = event.get("secret", "")
    else:
        token, expected, secret = "", "", ""

    if not token:
        return {"ok": False, "error": "Input 'token' is required"}
    if expected is None:
        return {"ok": False, "error": "Input 'expected' is required"}

    if secret:
        if "." not in token:
            return {"ok": True, "valid": False}
        part, sig = token.rsplit(".", 1)
        try:
            raw = base64.urlsafe_b64decode(part + "==")
        except Exception:
            return {"ok": True, "valid": False}
        expected_sig = hmac.new(secret.encode("utf-8") if isinstance(secret, str) else secret, raw, hashlib.sha256).hexdigest()
        valid = hmac.compare_digest(expected_sig, sig)
    else:
        valid = hmac.compare_digest(token, expected)

    return {"ok": True, "valid": valid}
