def handler(event):
    code = event.get("code") if isinstance(event, dict) else None
    if not code:
        return {"ok": False, "error": "code is required"}
    try:
        clean = str(code).strip().replace(" ", "")
        if not clean.isdigit():
            return {"ok": True, "result": False, "valid": False, "reason": "non-numeric"}
        digits = [int(c) for c in clean]
        if len(digits) == 12:
            upc_type = "UPC-A"
        elif len(digits) == 6:
            upc_type = "UPC-E"
            return {"ok": True, "result": True, "valid": True, "type": upc_type, "note": "UPC-E checksum requires expanded form"}
        else:
            return {"ok": True, "result": False, "valid": False, "reason": f"invalid length: {len(digits)} (must be 6 or 12)"}
        # UPC-A uses same algorithm as EAN with leading 0 parity
        odd = sum(digits[i] for i in range(0, 11, 2))
        even = sum(digits[i] for i in range(1, 11, 2))
        check = (10 - (odd * 3 + even) % 10) % 10
        valid = check == digits[11]
        return {"ok": True, "result": valid, "valid": valid, "type": upc_type, "check_digit": digits[-1]}
    except Exception as e:
        return {"ok": False, "error": str(e)}
