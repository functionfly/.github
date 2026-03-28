import urllib.request
import urllib.error
import ssl


SECURITY_HEADERS = {
    "strict-transport-security": {"name": "Strict-Transport-Security", "weight": 20},
    "content-security-policy": {"name": "Content-Security-Policy", "weight": 25},
    "x-frame-options": {"name": "X-Frame-Options", "weight": 15},
    "x-content-type-options": {"name": "X-Content-Type-Options", "weight": 10},
    "referrer-policy": {"name": "Referrer-Policy", "weight": 10},
    "permissions-policy": {"name": "Permissions-Policy", "weight": 10},
    "x-xss-protection": {"name": "X-XSS-Protection", "weight": 5},
    "cross-origin-opener-policy": {"name": "Cross-Origin-Opener-Policy", "weight": 5},
}


def get_grade(score):
    if score >= 90:
        return "A+"
    elif score >= 80:
        return "A"
    elif score >= 70:
        return "B"
    elif score >= 60:
        return "C"
    elif score >= 50:
        return "D"
    return "F"


def handler(event):
    """Analyze HTTP security headers."""
    try:
        url = event.get("url")
        if not url:
            return {"ok": False, "error": "url is required"}

        timeout = int(event.get("timeout", 10))

        ctx = ssl.create_default_context()
        ctx.check_hostname = False
        ctx.verify_mode = ssl.CERT_NONE

        try:
            req = urllib.request.Request(url, method="HEAD")
            with urllib.request.urlopen(req, timeout=timeout, context=ctx) as resp:
                headers = {k.lower(): v for k, v in resp.headers.items()}
        except urllib.error.HTTPError as e:
            headers = {k.lower(): v for k, v in e.headers.items()}

        present = []
        missing = []
        score = 0
        found_headers = {}

        for header_key, info in SECURITY_HEADERS.items():
            if header_key in headers:
                present.append(info["name"])
                score += info["weight"]
                found_headers[info["name"]] = headers[header_key]
            else:
                missing.append(info["name"])

        return {
            "ok": True,
            "score": score,
            "grade": get_grade(score),
            "present": present,
            "missing": missing,
            "headers": found_headers,
            "url": url,
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
