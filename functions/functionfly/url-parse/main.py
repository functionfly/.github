from urllib.parse import urlparse, parse_qs, parse_qsl


def handler(event):
    url = event.get("url")
    parse_query = event.get("parse_query", True)
    query_format = event.get("query_format", "object")  # "object" or "array"

    if not url:
        return {"ok": False, "error": "url is required"}

    if not isinstance(url, str):
        return {"ok": False, "error": "url must be a string"}

    try:
        parsed = urlparse(url)

        result = {
            "scheme": parsed.scheme,
            "netloc": parsed.netloc,
            "hostname": parsed.hostname,
            "port": parsed.port,
            "path": parsed.path,
            "params": parsed.params,
            "query": parsed.query,
            "fragment": parsed.fragment,
            "username": parsed.username,
            "password": parsed.password
        }

        # Add computed fields
        result["is_https"] = parsed.scheme == "https"
        result["is_http"] = parsed.scheme == "http"
        result["has_port"] = parsed.port is not None
        result["has_query"] = bool(parsed.query)
        result["has_fragment"] = bool(parsed.fragment)
        result["has_credentials"] = bool(parsed.username or parsed.password)
        result["is_absolute"] = bool(parsed.scheme and parsed.netloc)
        result["domain"] = parsed.hostname

        # Parse query string if requested
        if parse_query and parsed.query:
            if query_format == "object":
                query_params = parse_qs(parsed.query)
                # Convert single values to strings instead of lists
                for key, value in query_params.items():
                    if len(value) == 1:
                        query_params[key] = value[0]
            else:  # array format
                query_params = parse_qsl(parsed.query)

            result["query_params"] = query_params

        # Extract path segments
        if parsed.path:
            path_parts = parsed.path.strip('/').split('/') if parsed.path != '/' else []
            result["path_segments"] = path_parts
            result["path_depth"] = len(path_parts)

        return {
            "ok": True,
            "result": result,
            "original_url": url
        }

    except Exception as e:
        return {"ok": False, "error": f"failed to parse URL: {str(e)}"}