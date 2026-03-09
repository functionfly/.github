def handler(event):
    if isinstance(event, dict):
        token = event.get("token", event.get("jwt", ""))
        secret = event.get("secret", event.get("key", ""))
        algorithms = event.get("algorithms") or ["HS256"]
    else:
        token = secret = ""
        algorithms = ["HS256"]

    if not token:
        return {"ok": False, "error": "Input 'token' is required"}
    if not secret:
        return {"ok": False, "error": "Input 'secret' is required"}

    try:
        import jwt as pyjwt
    except ImportError:
        return {"ok": False, "error": "PyJWT is required; install with: pip install PyJWT"}

    try:
        payload = pyjwt.decode(token, secret, algorithms=algorithms)
        return {"ok": True, "payload": payload}
    except Exception as e:
        return {"ok": False, "error": str(e)}

