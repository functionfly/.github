import math

def _norm_cdf(x):
    return 0.5 * (1 + math.erf(x / math.sqrt(2)))

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
        option_type = str(event.get("option_type", "call")).lower()
        if T <= 0:
            intrinsic = max(S - K, 0) if option_type == "call" else max(K - S, 0)
            return {"ok": True, "result": round(intrinsic, 6)}
        d1 = (math.log(S / K) + (r + 0.5 * sigma ** 2) * T) / (sigma * math.sqrt(T))
        d2 = d1 - sigma * math.sqrt(T)
        if option_type == "call":
            price = S * _norm_cdf(d1) - K * math.exp(-r * T) * _norm_cdf(d2)
        else:
            price = K * math.exp(-r * T) * _norm_cdf(-d2) - S * _norm_cdf(-d1)
        return {"ok": True, "result": round(price, 6), "d1": round(d1, 6), "d2": round(d2, 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
