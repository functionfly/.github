def handler(event):
    if isinstance(event, dict):
        password = event.get("password", "")
        hash_str = event.get("hash", "")
    else:
        password = hash_str = ""
    if not password:
        return {"ok": False, "error": "Input password is required"}
    if not hash_str:
        return {"ok": False, "error": "Input hash is required"}
    try:
        import bcrypt
        pwd = password.encode("utf-8") if isinstance(password, str) else password
        h = hash_str.encode("utf-8") if isinstance(hash_str, str) else hash_str
        valid = bcrypt.checkpw(pwd, h)
        return {"ok": True, "valid": valid}
    except ImportError:
        return {"ok": False, "error": "bcrypt required: pip install bcrypt"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
