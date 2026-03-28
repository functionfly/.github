def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    periods = event.get("periods")
    payment = event.get("payment")
    pv = event.get("present_value")
    if periods is None or payment is None or pv is None:
        return {"ok": False, "error": "periods, payment, and present_value are required"}
    try:
        periods = float(periods)
        payment = float(payment)
        pv = float(pv)
        fv = float(event.get("future_value", 0))
        guess = float(event.get("guess", 0.01))
        rate = guess
        for _ in range(1000):
            if rate == 0:
                f = pv + payment * periods + fv
                df = 0
            else:
                f = (pv * (1 + rate) ** periods
                     + payment * ((1 + rate) ** periods - 1) / rate
                     + fv)
                df = (periods * pv * (1 + rate) ** (periods - 1)
                      + payment * (periods * (1 + rate) ** (periods - 1) * rate
                                   - ((1 + rate) ** periods - 1)) / rate ** 2)
            if abs(df) < 1e-12:
                break
            new_rate = rate - f / df
            if abs(new_rate - rate) < 1e-10:
                rate = new_rate
                break
            rate = new_rate
        return {"ok": True, "result": round(rate, 8)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
