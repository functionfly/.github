import hashlib
import os
import base64


def handler(event):
    password = event.get("password") if isinstance(event, dict) else None
    iterations = event.get("iterations", 600000)
    algorithm = event.get("algorithm", "sha256")
    dklen = event.get("dklen", 32)
    salt = event.get("salt")

    if not password:
        return {"ok": False, "error": "password is required"}

    SUPPORTED = ["sha1", "sha256", "sha384", "sha512", "sha3_256", "sha3_512"]
    if algorithm not in SUPPORTED:
        return {"ok": False, "error": f"algorithm must be one of: {', '.join(SUPPORTED)}"}

    try:
        pw = str(password).encode("utf-8")
        salt_bytes = base64.b64decode(str(salt)) if salt else os.urandom(16)
        dk = hashlib.pbkdf2_hmac(algorithm, pw, salt_bytes, int(iterations), dklen=int(dklen))
        salt_b64 = base64.b64encode(salt_bytes).decode("utf-8")
        dk_b64 = base64.b64encode(dk).decode("utf-8")
        encoded = f"$pbkdf2-{algorithm}$i={iterations}${salt_b64}${dk_b64}"
        return {"ok": True, "result": encoded, "salt": salt_b64, "hash": dk_b64}
    except Exception as e:
        return {"ok": False, "error": str(e)}
