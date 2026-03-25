def _ean_checksum(digits):
    total = 0
    for i, d in enumerate(digits[:-1]):
        total += d * (3 if i % 2 == 1 else 1)
    check = (10 - (total % 10)) % 10
    return check == digits[-1]


def handler(event):
    code = event.get("code") if isinstance(event, dict) else None
    if not code:
        return {"ok": False, "error": "code is required"}
    try:
        clean = str(code).strip().replace(" ", "")
        if not clean.isdigit():
            return {"ok": True, "result": False, "valid": False, "reason": "non-numeric"}
        digits = [int(c) for c in clean]
        if len(digits) == 13:
            ean_type = "EAN-13"
        elif len(digits) == 8:
            ean_type = "EAN-8"
        else:
            return {"ok": True, "result": False, "valid": False, "reason": f"invalid length: {len(digits)} (must be 8 or 13)"}
        valid = _ean_checksum(digits)
        return {"ok": True, "result": valid, "valid": valid, "type": ean_type, "check_digit": digits[-1]}
    except Exception as e:
        return {"ok": False, "error": str(e)}
