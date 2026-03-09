def handler(event):
    if isinstance(event, dict):
        origin = event.get("origin", "")
        allow_origin = event.get("allow_origin")
        methods = event.get("methods", ["GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"])
        allow_credentials = event.get("allow_credentials", False)
        max_age = event.get("max_age")
    else:
        origin, allow_origin, methods, allow_credentials, max_age = "", None, ["GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"], False, None

    if not origin:
        return {"ok": False, "error": "Input 'origin' is required"}

    if allow_origin == "*" or allow_origin is None:
        acao = "*"
    else:
        acao = allow_origin

    if isinstance(methods, list):
        methods_str = ", ".join(m.upper() for m in methods if m)
    else:
        methods_str = "GET, POST, PUT, DELETE, OPTIONS, PATCH"

    headers = {
        "Access-Control-Allow-Origin": acao,
        "Access-Control-Allow-Methods": methods_str,
        "Access-Control-Allow-Headers": "Content-Type, Authorization, X-Requested-With, Accept",
    }
    if allow_credentials and acao != "*":
        headers["Access-Control-Allow-Credentials"] = "true"
    if max_age is not None:
        headers["Access-Control-Max-Age"] = str(int(max_age))

    return {"ok": True, "headers": headers}
