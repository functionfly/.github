def handler(event):
    try:
        otp = event.get("otp", "") if isinstance(event, dict) else ""
        expected = event.get("expected", "") if isinstance(event, dict) else ""
        if not otp:
            return {"ok": False, "error": "otp is required"}
        if not expected:
            return {"ok": False, "error": "expected is required"}
        valid = otp == expected
        return {"ok": True, "valid": valid}
    except Exception as e:
        return {"ok": False, "error": str(e)}
