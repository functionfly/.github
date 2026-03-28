import math

def _norm_cdf(x):
    return 0.5 * (1 + math.erf(x / math.sqrt(2)))

def _norm_pdf(x):
    return math.exp(-0.5 * x ** 2) / math.sqrt(2 * math.pi)

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
            return {"ok": False, "error": "time_to_expiry must be positive"}
        d1 = (math.log(S / K) + (r + 0.5 * sigma ** 2) * T) / (sigma * math.sqrt(T))
        d2 = d1 - sigma * math.sqrt(T)
        if option_type == "call":
            delta = _norm_cdf(d1)
            rho = K * T * math.exp(-r * T) * _norm_cdf(d2)
        else:
            delta = _norm_cdf(d1) - 1
            rho = -K * T * math.exp(-r * T) * _norm_cdf(-d2)
        gamma = _norm_pdf(d1) / (S * sigma * math.sqrt(T))
        theta = ((-S * _norm_pdf(d1) * sigma / (2 * math.sqrt(T))
                  - r * K * math.exp(-r * T) * (_norm_cdf(d2) if option_type == "call" else _norm_cdf(-d2)))
                 / 365)
        vega = S * _norm_pdf(d1) * math.sqrt(T) / 100
        return {
            "ok": True,
            "result": {
                "delta": round(delta, 6),
                "gamma": round(gamma, 6),
                "theta": round(theta, 6),
                "vega": round(vega, 6),
                "rho": round(rho, 6)
            }
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
