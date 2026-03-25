import base64
import re


def handler(event):
    value = event.get("value") if isinstance(event, dict) else None

    if not value:
        return {"ok": False, "error": "value is required"}
    if not isinstance(value, str):
        return {"ok": False, "error": "value must be a string"}

    value = value.strip()
    parts = value.split(" ", 1)
    if len(parts) < 2:
        return {"ok": False, "error": "Invalid Authorization header format"}

    scheme = parts[0].strip()
    credentials = parts[1].strip()
    scheme_lower = scheme.lower()

    result = {
        "ok": True,
        "scheme": scheme,
        "credentials": credentials,
    }

    if scheme_lower == "basic":
        try:
            decoded = base64.b64decode(credentials).decode("utf-8")
            if ":" in decoded:
                username, password = decoded.split(":", 1)
                result["username"] = username
                result["password"] = password
        except Exception:
            pass

    elif scheme_lower == "bearer":
        result["token"] = credentials

    elif scheme_lower == "digest":
        digest_params = {}
        for match in re.finditer(r'(\w+)=["\']?([^"\'\\,]+)["\']?', credentials):
            digest_params[match.group(1)] = match.group(2).strip()
        result["params"] = digest_params

    return result
