import re


def handler(event):
    """Generate a git tag name from a version string."""
    try:
        version = event.get("version")
        if not version:
            return {"ok": False, "error": "version is required"}

        version = str(version).strip().lstrip("v")
        prefix = event.get("prefix", "v")
        fmt = event.get("format")

        if fmt:
            tag = fmt.format(prefix=prefix, version=version)
        else:
            tag = f"{prefix}{version}"

        # Validate tag name (no spaces, no special chars except . - _)
        if not re.match(r'^[a-zA-Z0-9._\-/]+$', tag):
            return {"ok": False, "error": f"Invalid tag name: {tag}"}

        return {"ok": True, "result": tag, "tag": tag}
    except Exception as e:
        return {"ok": False, "error": str(e)}
