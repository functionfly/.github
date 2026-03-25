def handler(event):
    number = event.get("number") if isinstance(event, dict) else None
    if not number:
        return {"ok": False, "error": "number is required"}
    try:
        digits = [int(c) for c in str(number).replace(" ", "").replace("-", "") if c.isdigit()]
        if len(digits) < 12 or len(digits) > 19:
            return {"ok": True, "result": False, "valid": False, "reason": "invalid length"}
        total = 0
        for i, d in enumerate(reversed(digits)):
            if i % 2 == 1:
                d *= 2
                if d > 9:
                    d -= 9
            total += d
        valid = total % 10 == 0
        return {"ok": True, "result": valid, "valid": valid, "length": len(digits)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
