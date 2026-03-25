def handler(event):
    password = event.get("password") if isinstance(event, dict) else None
    rounds = event.get("rounds", 12)

    if not password:
        return {"ok": False, "error": "password is required"}
    try:
        import bcrypt
        pw = str(password).encode("utf-8")
        salt = bcrypt.gensalt(rounds=int(rounds))
        hashed = bcrypt.hashpw(pw, salt)
        return {"ok": True, "result": hashed.decode("utf-8"), "rounds": int(rounds)}
    except ImportError:
        return {"ok": False, "error": "bcrypt library is not installed. Install with: pip install bcrypt"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
