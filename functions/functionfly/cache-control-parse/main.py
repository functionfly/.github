import re


def handler(event):
    value = event.get("value") if isinstance(event, dict) else None

    if not value:
        return {"ok": False, "error": "value is required"}
    if not isinstance(value, str):
        return {"ok": False, "error": "value must be a string"}

    directives = {}
    for part in value.split(","):
        part = part.strip()
        if "=" in part:
            k, v = part.split("=", 1)
            k = k.strip().lower()
            v = v.strip().strip('"')
            try:
                directives[k] = int(v)
            except ValueError:
                directives[k] = v
        elif part:
            directives[part.lower()] = True

    return {
        "ok": True,
        "value": value,
        "directives": directives,
        "max_age": directives.get("max-age"),
        "s_maxage": directives.get("s-maxage"),
        "no_cache": directives.get("no-cache", False),
        "no_store": directives.get("no-store", False),
        "must_revalidate": directives.get("must-revalidate", False),
        "public": directives.get("public", False),
        "private": directives.get("private", False),
        "immutable": directives.get("immutable", False),
    }
