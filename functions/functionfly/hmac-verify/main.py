import base64
import hmac
import hashlib


def handler(event):
    if isinstance(event, dict):
        message = event.get("message", event.get("data", ""))
        secret = event.get("secret", event.get("key", ""))
        signature = event.get("signature", "")
        algorithm = event.get("algorithm", "sha256").lower().replace("-", "")
    else:
        message = secret = signature = ""
        algorithm = "sha256"

    if not secret:
        return {"ok": False, "error": "Input 'secret' is required"}
    if not signature:
        return {"ok": False, "error": "Input 'signature' is required"}

    msg_bytes = message.encode("utf-8") if isinstance(message, str) else message
    key_bytes = secret.encode("utf-8") if isinstance(secret, str) else secret

    sig_str = str(signature).strip()
    try:
        if len(sig_str) == 64 and all(c in "0123456789abcdef" for c in sig_str.lower()):
            expected = bytes.fromhex(sig_str)
        else:
            expected = base64.b64decode(sig_str, validate=True)
    except Exception as e:
        return {"ok": False, "error": f"Invalid signature format: {e}"}

    alg_map = {"sha256": hashlib.sha256, "sha512": hashlib.sha512, "sha1": hashlib.sha1, "md5": hashlib.md5}
    digest = alg_map.get(algorithm, hashlib.sha256)
    computed = hmac.new(key_bytes, msg_bytes, digest).digest()
    valid = hmac.compare_digest(computed, expected)
    return {"ok": True, "valid": valid}

