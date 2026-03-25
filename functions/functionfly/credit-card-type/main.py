import re


CARD_PATTERNS = [
    ("amex",       r"^3[47][0-9]{13}$",                 "American Express"),
    ("visa",       r"^4[0-9]{12}(?:[0-9]{3})?$",        "Visa"),
    ("mastercard", r"^5[1-5][0-9]{14}$",                 "Mastercard"),
    ("mastercard", r"^2(?:2[2-9][1-9]|2[3-9][0-9]|[3-6][0-9]{2}|7[01][0-9]|720)[0-9]{12}$", "Mastercard"),
    ("discover",   r"^6(?:011|5[0-9]{2})[0-9]{12}$",    "Discover"),
    ("jcb",        r"^(?:2131|1800|35\d{3})\d{11}$",    "JCB"),
    ("diners",     r"^3(?:0[0-5]|[68][0-9])[0-9]{11}$", "Diners Club"),
    ("unionpay",   r"^62[0-9]{14,17}$",                  "UnionPay"),
    ("maestro",    r"^(?:5018|5020|5038|6304|6759|6761|6763)[0-9]{8,15}$", "Maestro"),
]


def handler(event):
    number = event.get("number") if isinstance(event, dict) else None
    if not number:
        return {"ok": False, "error": "number is required"}
    try:
        clean = str(number).replace(" ", "").replace("-", "")
        for card_type, pattern, name in CARD_PATTERNS:
            if re.match(pattern, clean):
                return {"ok": True, "result": card_type, "type": card_type, "name": name, "length": len(clean)}
        return {"ok": True, "result": "unknown", "type": "unknown", "name": "Unknown", "length": len(clean)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
