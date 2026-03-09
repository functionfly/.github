from urllib.parse import urlunparse, urlencode


def handler(event):
    components = event.get("components", {})
    query_params = event.get("query_params")

    if not isinstance(components, dict):
        return {"ok": False, "error": "components must be an object"}

    try:
        # Extract URL components
        scheme = components.get("scheme", "")
        netloc = components.get("netloc", "")
        path = components.get("path", "")
        params = components.get("params", "")
        query = components.get("query", "")
        fragment = components.get("fragment", "")

        # Handle hostname/port/netloc construction
        hostname = components.get("hostname")
        port = components.get("port")
        username = components.get("username")
        password = components.get("password")

        if hostname and not netloc:
            netloc = hostname
            if port:
                netloc = f"{hostname}:{port}"
            if username:
                if password:
                    netloc = f"{username}:{password}@{netloc}"
                else:
                    netloc = f"{username}@{netloc}"

        # Handle query parameters
        if query_params:
            if isinstance(query_params, dict):
                query = urlencode(query_params)
            elif isinstance(query_params, list):
                # List of (key, value) tuples
                query = urlencode(query_params)

        # Build URL using urlunparse
        url = urlunparse((scheme, netloc, path, params, query, fragment))

        result = {
            "url": url,
            "components_used": {
                "scheme": scheme,
                "netloc": netloc,
                "path": path,
                "params": params,
                "query": query,
                "fragment": fragment
            },
            "has_scheme": bool(scheme),
            "has_netloc": bool(netloc),
            "has_path": bool(path),
            "has_query": bool(query),
            "has_fragment": bool(fragment),
            "is_absolute": bool(scheme and netloc)
        }

        return {
            "ok": True,
            "result": result
        }

    except Exception as e:
        return {"ok": False, "error": f"failed to build URL: {str(e)}"}