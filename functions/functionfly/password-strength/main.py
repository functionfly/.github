import re

def handler(event):
    if isinstance(event, dict):
        password = event.get("password", "")
    else:
        password = ""
    if password is None:
        return {"ok": False, "error": "Input 'password' is required"}
    s = str(password)
    checks = {
        "length_ok": len(s) >= 8,
        "has_uppercase": bool(re.search(r"[A-Z]", s)),
        "has_lowercase": bool(re.search(r"[a-z]", s)),
        "has_digit": bool(re.search(r"\d", s)),
        "has_special": bool(re.search(r"[!@#$%^&*()_+\-=\[\]{};':\"\\|,.<>/?]", s)),
        "length_12_plus": len(s) >= 12,
    }
    score = sum(1 for v in checks.values() if v)
    label = "weak" if score <= 2 else "fair" if score <= 4 else "good" if score <= 5 else "strong"
    return {"ok": True, "score": score, "label": label, "checks": checks}
