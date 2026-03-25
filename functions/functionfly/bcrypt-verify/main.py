def handler(event):
    password = event.get("password") if isinstance(event, dict) else None
    hash_ = event.get("hash")

    if not password:
        return {"ok": False, "error": "password is required"}
    if not hash_:
        return {"ok": False, "error": "hash is required"}
    try:
        import bcrypt
        pw = str(password).encode("utf-8")
        h = str(hash_).encode("utf-8")
        match = bcrypt.checkpw(pw, h)
        return {"ok": True, "result": match}
    except ImportError:
        return {"ok": False, "error": "bcrypt library is not installed. Install with: pip install bcrypt"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
