import hashlib
import base64


def handler(event):
    password = event.get("password") if isinstance(event, dict) else None
    salt = event.get("salt")
    n = event.get("n", 16384)
    r = event.get("r", 8)
    p = event.get("p", 1)
    length = event.get("length", 32)
    output = event.get("output", "hex")

    if not password:
        return {"ok": False, "error": "password is required"}
    if not salt:
        return {"ok": False, "error": "salt is required"}

    try:
        try:
            salt_bytes = bytes.fromhex(str(salt))
        except ValueError:
            salt_bytes = base64.b64decode(str(salt))

        dk = hashlib.scrypt(str(password).encode("utf-8"), salt=salt_bytes, n=int(n), r=int(r), p=int(p), dklen=int(length))
        result = base64.b64encode(dk).decode("utf-8") if output == "base64" else dk.hex()
        return {"ok": True, "result": result, "n": int(n), "r": int(r), "p": int(p), "length": int(length)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
