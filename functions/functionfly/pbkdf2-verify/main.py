import hashlib
import base64
import hmac


def handler(event):
    password = event.get("password") if isinstance(event, dict) else None
    hash_ = event.get("hash")

    if not password:
        return {"ok": False, "error": "password is required"}
    if not hash_:
        return {"ok": False, "error": "hash is required"}
    try:
        # Format: $pbkdf2-{algorithm}$i={iterations}${salt_b64}${hash_b64}
        parts = str(hash_).split("$")
        if len(parts) != 5 or not parts[1].startswith("pbkdf2-"):
            return {"ok": False, "error": "invalid PBKDF2 hash format"}
        algorithm = parts[1].replace("pbkdf2-", "")
        iterations = int(parts[2].replace("i=", ""))
        salt_bytes = base64.b64decode(parts[3])
        stored_hash = base64.b64decode(parts[4])
        dklen = len(stored_hash)
        pw = str(password).encode("utf-8")
        dk = hashlib.pbkdf2_hmac(algorithm, pw, salt_bytes, iterations, dklen=dklen)
        match = hmac.compare_digest(dk, stored_hash)
        return {"ok": True, "result": match}
    except Exception as e:
        return {"ok": False, "error": str(e)}
