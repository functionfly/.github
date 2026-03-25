def handler(event):
    headers = event.get("headers", {}) if isinstance(event, dict) else {}
    query_params = event.get("query_params", {}) if isinstance(event, dict) else {}
    header_name = event.get("header_name", "X-API-Key")
    query_param_name = event.get("query_param_name", "api_key")

    if not isinstance(headers, dict):
        return {"ok": False, "error": "headers must be an object"}
    if not isinstance(query_params, dict):
        return {"ok": False, "error": "query_params must be an object"}

    # Search headers (case-insensitive)
    headers_lower = {k.lower(): v for k, v in headers.items()}
    header_key = header_name.lower()

    api_key = headers_lower.get(header_key)
    if api_key:
        return {"ok": True, "api_key": api_key, "source": "header", "header_name": header_name}

    # Check Authorization: ApiKey <key>
    auth = headers_lower.get("authorization", "")
    if auth.lower().startswith("apikey "):
        token = auth[7:].strip()
        if token:
            return {"ok": True, "api_key": token, "source": "authorization"}

    # Search query parameters
    api_key = query_params.get(query_param_name)
    if api_key:
        return {"ok": True, "api_key": api_key, "source": "query", "param_name": query_param_name}

    return {"ok": False, "error": "API key not found in headers or query parameters"}
