import math

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    required = ["spot_price", "strike_price", "time_to_expiry", "risk_free_rate", "volatility"]
    for f in required:
        if event.get(f) is None:
            return {"ok": False, "error": f"{f} is required"}
    try:
        S = float(event["spot_price"])
        K = float(event["strike_price"])
        T = float(event["time_to_expiry"])
        r = float(event["risk_free_rate"])
        sigma = float(event["volatility"])
        N = int(event.get("steps", 100))
        option_type = str(event.get("option_type", "call")).lower()
        # Binomial tree (CRR model)
        dt = T / N
        u = math.exp(sigma * math.sqrt(dt))
        d = 1 / u
        p = (math.exp(r * dt) - d) / (u - d)
        # Terminal payoffs
        prices = [S * u ** (N - 2 * j) for j in range(N + 1)]
        if option_type == "call":
            values = [max(price - K, 0) for price in prices]
        else:
            values = [max(K - price, 0) for price in prices]
        # Backward induction
        discount = math.exp(-r * dt)
        for i in range(N - 1, -1, -1):
            values = [discount * (p * values[j] + (1 - p) * values[j + 1]) for j in range(i + 1)]
        return {"ok": True, "result": round(values[0], 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
