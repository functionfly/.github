import re

try:
    import dns.resolver
    HAS_DNS = True
except ImportError:
    HAS_DNS = False


def _check_mx(domain):
    if not HAS_DNS:
        return None
    try:
        mx_records = dns.resolver.resolve(domain, "MX")
        return len(mx_records) > 0
    except Exception:
        return False


def handler(event):
    """
    Validate email format and basic deliverability hints.

    Input:
        - email: Email address to validate (required)

    Returns:
        - ok: True if format is valid
        - valid: Same as ok
        - format_valid: True if pattern matches
        - has_mx: True if domain has MX record
        - error: Message if invalid
    """
    if isinstance(event, dict):
        email = event.get("email", event.get("address", ""))
    else:
        email = str(event) if event else ""

    if not email or not str(email).strip():
        return {"ok": False, "valid": False, "format_valid": False, "error": "Input 'email' is required"}

    email = str(email).strip().lower()

    pattern = r"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*\.[a-zA-Z]{2,}$"
    if not re.match(pattern, email):
        return {
            "ok": False,
            "valid": False,
            "format_valid": False,
            "error": "Invalid email format",
        }

    local, _, domain = email.partition("@")
    if ".." in email or local.startswith(".") or local.endswith("."):
        return {
            "ok": False,
            "valid": False,
            "format_valid": False,
            "error": "Invalid email format",
        }

    if len(local) > 64 or len(domain) > 255 or len(email) > 254:
        return {
            "ok": False,
            "valid": False,
            "format_valid": False,
            "error": "Email too long",
        }

    has_mx = _check_mx(domain)

    return {
        "ok": True,
        "valid": True,
        "format_valid": True,
        "email": email,
        "local": local,
        "domain": domain,
        "has_mx": has_mx,
    }
