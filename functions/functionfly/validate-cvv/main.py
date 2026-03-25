def handler(event):
    cvv = event.get("cvv") if isinstance(event, dict) else None
    card_type = event.get("card_type", "unknown")
    if not cvv:
        return {"ok": False, "error": "cvv is required"}
    try:
        cvv_str = str(cvv).strip()
        if not cvv_str.isdigit():
            return {"ok": True, "result": False, "valid": False, "reason": "CVV must contain only digits"}
        expected_length = 4 if card_type == "amex" else 3
        valid = len(cvv_str) == expected_length
        return {"ok": True, "result": valid, "valid": valid, "length": len(cvv_str), "expected_length": expected_length}
    except Exception as e:
        return {"ok": False, "error": str(e)}
