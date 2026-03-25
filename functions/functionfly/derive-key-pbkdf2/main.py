import hashlib
import base64


def handler(event):
    password = event.get("password") if isinstance(event, dict) else None
    salt = event.get("salt")
    iterations = event.get("iterations", 600000)
    length = event.get("length", 32)
    algorithm = event.get("algorithm", "sha256")
    output = event.get("output", "hex")

    if not password:
        return {"ok": False, "error": "password is required"}
    if not salt:
        return {"ok": False, "error": "salt is required (use generate-salt to create one)"}

    try:
        try:
            salt_bytes = bytes.fromhex(str(salt))
        except ValueError:
            salt_bytes = base64.b64decode(str(salt))

        dk = hashlib.pbkdf2_hmac(algorithm, str(password).encode("utf-8"), salt_bytes, int(iterations), dklen=int(length))
        result = base64.b64encode(dk).decode("utf-8") if output == "base64" else dk.hex()
        return {"ok": True, "result": result, "algorithm": algorithm, "iterations": int(iterations), "length": int(length)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
