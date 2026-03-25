import hashlib
import base64
import os


def handler(event):
    """Derive multiple sub-keys from a single master password using PBKDF2."""
    password = event.get("password") if isinstance(event, dict) else None
    purposes = event.get("purposes", ["encryption", "authentication"])
    iterations = event.get("iterations", 600000)
    key_length = event.get("key_length", 32)
    algorithm = event.get("algorithm", "sha256")

    if not password:
        return {"ok": False, "error": "password is required"}

    try:
        pw = str(password).encode("utf-8")
        derived = {}
        for purpose in purposes:
            # Deterministic salt per purpose using a fixed prefix
            salt = hashlib.sha256(f"purpose:{purpose}".encode()).digest()
            dk = hashlib.pbkdf2_hmac(algorithm, pw, salt, int(iterations), dklen=int(key_length))
            derived[str(purpose)] = dk.hex()
        return {"ok": True, "result": derived, "algorithm": algorithm, "key_length": int(key_length)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
