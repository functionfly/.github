def handler(event):
    password = event.get("password") if isinstance(event, dict) else None
    time_cost = event.get("time_cost", 3)
    memory_cost = event.get("memory_cost", 65536)
    parallelism = event.get("parallelism", 4)
    hash_len = event.get("hash_len", 32)

    if not password:
        return {"ok": False, "error": "password is required"}

    try:
        from argon2 import PasswordHasher
        ph = PasswordHasher(
            time_cost=int(time_cost),
            memory_cost=int(memory_cost),
            parallelism=int(parallelism),
            hash_len=int(hash_len),
        )
        result = ph.hash(str(password))
        return {"ok": True, "result": result}
    except ImportError:
        pass

    try:
        import hashlib, os, base64
        salt = os.urandom(16)
        dk = hashlib.scrypt(
            str(password).encode("utf-8"), salt=salt,
            n=16384, r=8, p=1, dklen=32
        )
        encoded = f"$scrypt-fallback${base64.b64encode(salt).decode()}${base64.b64encode(dk).decode()}"
        return {"ok": True, "result": encoded, "note": "argon2-cffi not installed; used scrypt fallback"}
    except Exception as e:
        return {"ok": False, "error": f"argon2-cffi not installed and fallback failed: {e}"}
