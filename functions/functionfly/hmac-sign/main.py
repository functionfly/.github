import hmac
import hashlib


def handler(event):
    if isinstance(event, dict):
        message = event.get("message", event.get("data", ""))
        secret = event.get("secret", event.get("key", ""))
        algorithm = event.get("algorithm", "sha256").lower().replace("-", "")
    else:
        message = secret = ""
        algorithm = "sha256"

    if not secret:
        return {"ok": False, "error": "Input 'secret' is required"}

    msg_bytes = message.encode("utf-8") if isinstance(message, str) else message
    key_bytes = secret.encode("utf-8") if isinstance(secret, str) else secret

    alg_map = {"sha256": hashlib.sha256, "sha512": hashlib.sha512, "sha1": hashlib.sha1, "md5": hashlib.md5}
    digest = alg_map.get(algorithm, hashlib.sha256)
    sig = hmac.new(key_bytes, msg_bytes, digest).digest()
    import base64
    b64 = base64.b64encode(sig).decode("ascii")
    hex_sig = sig.hex()
    return {"ok": True, "signature": b64, "hex": hex_sig}

