def handler(event):
    if isinstance(event, dict):
        password = event.get("password", "")
        rounds = event.get("rounds", 12)
    else:
        password, rounds = "", 12
    if not password:
        return {"ok": False, "error": "Input password is required"}
    try:
        rounds = max(4, min(31, int(rounds)))
    except (TypeError, ValueError):
        rounds = 12
    try:
        import bcrypt
        data = password.encode("utf-8") if isinstance(password, str) else password
        hashed = bcrypt.hashpw(data, bcrypt.gensalt(rounds=rounds))
        return {"ok": True, "hash": hashed.decode("utf-8")}
    except ImportError:
        return {"ok": False, "error": "bcrypt required: pip install bcrypt"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
