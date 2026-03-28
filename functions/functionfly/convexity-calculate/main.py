def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    required = ["face_value", "coupon_rate", "years_to_maturity", "yield_to_maturity"]
    for f in required:
        if event.get(f) is None:
            return {"ok": False, "error": f"{f} is required"}
    try:
        F = float(event["face_value"])
        c_rate = float(event["coupon_rate"])
        years = float(event["years_to_maturity"])
        ytm = float(event["yield_to_maturity"])
        freq = int(event.get("periods_per_year", 2))
        n = int(years * freq)
        coupon = F * c_rate / freq
        r = ytm / freq
        if r == 0:
            price = coupon * n + F
            convexity_sum = sum(t * (t + 1) * coupon for t in range(1, n + 1)) + n * (n + 1) * F
        else:
            price = sum(coupon / (1 + r) ** t for t in range(1, n + 1)) + F / (1 + r) ** n
            convexity_sum = (sum(t * (t + 1) * coupon / (1 + r) ** (t + 2) for t in range(1, n + 1))
                             + n * (n + 1) * F / (1 + r) ** (n + 2))
        convexity = convexity_sum / (price * freq ** 2)
        return {"ok": True, "result": round(convexity, 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
