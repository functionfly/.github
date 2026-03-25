import re


def luhn_check(number):
    digits = [int(d) for d in str(number)]
    digits.reverse()
    total = 0
    for i, d in enumerate(digits):
        if i % 2 == 1:
            d *= 2
            if d > 9:
                d -= 9
        total += d
    return total % 10 == 0


CARD_PATTERNS = {
    "visa": re.compile(r'^4[0-9]{12}(?:[0-9]{3})?$'),
    "mastercard": re.compile(r'^5[1-5][0-9]{14}$'),
    "amex": re.compile(r'^3[47][0-9]{13}$'),
    "discover": re.compile(r'^6(?:011|5[0-9]{2})[0-9]{12}$'),
    "diners": re.compile(r'^3(?:0[0-5]|[68][0-9])[0-9]{11}$'),
    "jcb": re.compile(r'^(?:2131|1800|35\d{3})\d{11}$'),
}


def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    if value is None:
        return {"ok": False, "error": "value is required"}

    # Strip spaces and dashes
    clean = re.sub(r'[\s\-]', '', str(value))
    if not clean.isdigit():
        return {"ok": True, "value": value, "result": False, "card_type": None}

    luhn_valid = luhn_check(clean)
    card_type = None
    for name, pattern in CARD_PATTERNS.items():
        if pattern.match(clean):
            card_type = name
            break

    result = luhn_valid and len(clean) >= 13 and len(clean) <= 19
    return {"ok": True, "value": value, "result": result, "card_type": card_type, "luhn_valid": luhn_valid}
