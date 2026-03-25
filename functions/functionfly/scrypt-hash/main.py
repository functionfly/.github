import hashlib
import os
import base64


def handler(event):
    password = event.get("password") if isinstance(event, dict) else None
    n = event.get("n", 16384)
    r = event.get("r", 8)
    p = event.get("p", 1)
    dklen = event.get("dklen", 32)
    salt = event.get("salt")

    if not password:
        return {"ok": False, "error": "password is required"}
    try:
        pw = str(password).encode("utf-8")
        if salt:
            salt_bytes = base64.b64decode(str(salt))
        else:
            salt_bytes = os.urandom(16)
        dk = hashlib.scrypt(pw, salt=salt_bytes, n=int(n), r=int(r), p=int(p), dklen=int(dklen))
        salt_b64 = base64.b64encode(salt_bytes).decode("utf-8")
        dk_b64 = base64.b64encode(dk).decode("utf-8")
        encoded = f"$scrypt$n={n}$r={r}$p={p}${salt_b64}${dk_b64}"
        return {"ok": True, "result": encoded, "salt": salt_b64, "hash": dk_b64}
    except Exception as e:
        return {"ok": False, "error": str(e)}
