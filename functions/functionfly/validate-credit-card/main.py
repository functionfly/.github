import re


CARD_PATTERNS = [
    ("amex",       r"^3[47][0-9]{13}$"),
    ("visa",       r"^4[0-9]{12}(?:[0-9]{3})?$"),
    ("mastercard", r"^5[1-5][0-9]{14}$"),
    ("mastercard", r"^2(?:2[2-9][1-9]|2[3-9][0-9]|[3-6][0-9]{2}|7[01][0-9]|720)[0-9]{12}$"),
    ("discover",   r"^6(?:011|5[0-9]{2})[0-9]{12}$"),
    ("jcb",        r"^(?:2131|1800|35\d{3})\d{11}$"),
    ("diners",     r"^3(?:0[0-5]|[68][0-9])[0-9]{11}$"),
    ("unionpay",   r"^62[0-9]{14,17}$"),
]

def _luhn(digits):
    total = 0
    for i, d in enumerate(reversed(digits)):
        d = int(d)
        if i % 2 == 1:
            d *= 2
            if d > 9: d -= 9
        total += d
    return total % 10 == 0

def handler(event):
    number = event.get("number") if isinstance(event, dict) else None
    if not number:
        return {"ok": False, "error": "number is required"}
    try:
        clean = str(number).replace(" ", "").replace("-", "")
        if not clean.isdigit():
            return {"ok": True, "result": False, "valid": False, "reason": "non-numeric characters"}
        luhn_ok = _luhn(clean)
        card_type = "unknown"
        for t, p in CARD_PATTERNS:
            if re.match(p, clean):
                card_type = t; break
        return {"ok": True, "result": luhn_ok, "valid": luhn_ok, "card_type": card_type, "length": len(clean)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
