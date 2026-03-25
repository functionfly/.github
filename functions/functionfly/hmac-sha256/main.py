import hmac
import hashlib
import base64


def handler(event):
    message = event.get("message") if isinstance(event, dict) else None
    secret = event.get("secret")
    output = event.get("output", "hex")

    if message is None:
        return {"ok": False, "error": "message is required"}
    if not secret:
        return {"ok": False, "error": "secret is required"}

    try:
        if isinstance(message, (dict, list)):
            import json
            msg_bytes = json.dumps(message).encode("utf-8")
        else:
            msg_bytes = str(message).encode("utf-8")
        key_bytes = str(secret).encode("utf-8")
        h = hmac.new(key_bytes, msg_bytes, hashlib.sha256)
        digest = h.digest()
        if output == "base64":
            result = base64.b64encode(digest).decode("utf-8")
        elif output == "base64url":
            result = base64.urlsafe_b64encode(digest).rstrip(b"=").decode("utf-8")
        else:
            result = h.hexdigest()
        return {"ok": True, "result": result, "algorithm": "HMAC-SHA256", "output_format": output}
    except Exception as e:
        return {"ok": False, "error": str(e)}
