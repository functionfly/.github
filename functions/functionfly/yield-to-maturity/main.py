def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    required = ["face_value", "coupon_rate", "years_to_maturity", "market_price"]
    for f in required:
        if event.get(f) is None:
            return {"ok": False, "error": f"{f} is required"}
    try:
        F = float(event["face_value"])
        c_rate = float(event["coupon_rate"])
        years = float(event["years_to_maturity"])
        P = float(event["market_price"])
        freq = int(event.get("periods_per_year", 2))
        n = int(years * freq)
        coupon = F * c_rate / freq
        # Newton-Raphson to find YTM
        ytm = c_rate  # initial guess
        for _ in range(1000):
            r = ytm / freq
            price = sum(coupon / (1 + r) ** t for t in range(1, n + 1)) + F / (1 + r) ** n
            dprice = sum(-t * coupon / (1 + r) ** (t + 1) for t in range(1, n + 1)) - n * F / (1 + r) ** (n + 1)
            dprice /= freq
            diff = price - P
            if abs(dprice) < 1e-12:
                break
            new_ytm = ytm - diff / dprice
            if abs(new_ytm - ytm) < 1e-10:
                ytm = new_ytm
                break
            ytm = new_ytm
        return {"ok": True, "result": round(ytm, 8)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
