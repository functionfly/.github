from urllib.parse import urljoin, urlparse


def handler(event):
    base_url = event.get("base_url")
    relative_url = event.get("relative_url")

    if not base_url:
        return {"ok": False, "error": "base_url is required"}

    if not relative_url:
        return {"ok": False, "error": "relative_url is required"}

    if not isinstance(base_url, str) or not isinstance(relative_url, str):
        return {"ok": False, "error": "both base_url and relative_url must be strings"}

    try:
        # Use urljoin to resolve relative URL against base URL
        resolved_url = urljoin(base_url, relative_url)

        # Parse both URLs for analysis
        base_parsed = urlparse(base_url)
        relative_parsed = urlparse(relative_url)
        resolved_parsed = urlparse(resolved_url)

        result = {
            "resolved_url": resolved_url,
            "base_url": base_url,
            "relative_url": relative_url,
            "was_modified": resolved_url != relative_url,
            "is_absolute_result": bool(resolved_parsed.scheme and resolved_parsed.netloc),
            "scheme_preserved": resolved_parsed.scheme == base_parsed.scheme,
            "host_preserved": resolved_parsed.hostname == base_parsed.hostname
        }

        # Add change analysis
        changes = []
        if resolved_parsed.scheme != relative_parsed.scheme and relative_parsed.scheme:
            changes.append("scheme was preserved from relative URL")
        elif resolved_parsed.scheme != base_parsed.scheme:
            changes.append("scheme was inherited from base URL")

        if resolved_parsed.netloc != relative_parsed.netloc and relative_parsed.netloc:
            changes.append("netloc was preserved from relative URL")
        elif resolved_parsed.netloc != base_parsed.netloc:
            changes.append("netloc was inherited from base URL")

        if resolved_parsed.path != relative_parsed.path:
            if relative_parsed.path.startswith('/'):
                changes.append("absolute path from relative URL replaced base path")
            else:
                changes.append("relative path was resolved against base path")

        result["changes"] = changes

        return {
            "ok": True,
            "result": result
        }

    except Exception as e:
        return {"ok": False, "error": f"failed to resolve URL: {str(e)}"}