def handler(event):
    if isinstance(event, dict):
        payload = event.get("payload", event.get("data", {}))
        secret = event.get("secret", event.get("key", ""))
        algorithm = event.get("algorithm", "HS256")
        headers = event.get("headers")
    else:
        payload = {}
        secret = ""
        algorithm = "HS256"
        headers = None

    if not isinstance(payload, dict):
        return {"ok": False, "error": "Input 'payload' must be an object"}
    if not secret:
        return {"ok": False, "error": "Input 'secret' is required"}

    try:
        import jwt as pyjwt
    except ImportError:
        return {"ok": False, "error": "PyJWT is required; install with: pip install PyJWT"}

    try:
        opts = {"algorithm": algorithm}
        if headers:
            opts["headers"] = headers
        token = pyjwt.encode(payload, secret, **opts)
        if hasattr(token, "decode"):
            token = token.decode("utf-8")
        return {"ok": True, "token": token}
    except Exception as e:
        return {"ok": False, "error": str(e)}

