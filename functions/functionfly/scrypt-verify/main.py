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
        # Parse format: $scrypt$n=N$r=R$p=P$salt_b64$hash_b64
        parts = str(hash_).split("$")
        if len(parts) != 7 or parts[1] != "scrypt":
            return {"ok": False, "error": "invalid scrypt hash format"}
        params = {}
        for p in parts[2:5]:
            k, v = p.split("=")
            params[k] = int(v)
        salt_bytes = base64.b64decode(parts[5])
        stored_hash = base64.b64decode(parts[6])
        dklen = len(stored_hash)
        pw = str(password).encode("utf-8")
        dk = hashlib.scrypt(pw, salt=salt_bytes, n=params["n"], r=params["r"], p=params["p"], dklen=dklen)
        match = hmac.compare_digest(dk, stored_hash)
        return {"ok": True, "result": match}
    except Exception as e:
        return {"ok": False, "error": str(e)}
