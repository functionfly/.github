import math

def _norm_cdf(x):
    return 0.5 * (1 + math.erf(x / math.sqrt(2)))

def _norm_pdf(x):
    return math.exp(-0.5 * x ** 2) / math.sqrt(2 * math.pi)

def _bs_price(S, K, T, r, sigma, option_type):
    if T <= 0 or sigma <= 0:
        return max(S - K, 0) if option_type == "call" else max(K - S, 0)
    d1 = (math.log(S / K) + (r + 0.5 * sigma ** 2) * T) / (sigma * math.sqrt(T))
    d2 = d1 - sigma * math.sqrt(T)
    if option_type == "call":
        return S * _norm_cdf(d1) - K * math.exp(-r * T) * _norm_cdf(d2)
    else:
        return K * math.exp(-r * T) * _norm_cdf(-d2) - S * _norm_cdf(-d1)

def _vega(S, K, T, r, sigma):
    if T <= 0 or sigma <= 0:
        return 0
    d1 = (math.log(S / K) + (r + 0.5 * sigma ** 2) * T) / (sigma * math.sqrt(T))
    return S * _norm_pdf(d1) * math.sqrt(T)

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    required = ["option_price", "spot_price", "strike_price", "time_to_expiry", "risk_free_rate"]
    for f in required:
        if event.get(f) is None:
            return {"ok": False, "error": f"{f} is required"}
    try:
        market_price = float(event["option_price"])
        S = float(event["spot_price"])
        K = float(event["strike_price"])
        T = float(event["time_to_expiry"])
        r = float(event["risk_free_rate"])
        option_type = str(event.get("option_type", "call")).lower()
        sigma = 0.2  # initial guess
        for _ in range(100):
            price = _bs_price(S, K, T, r, sigma, option_type)
            v = _vega(S, K, T, r, sigma)
            if abs(v) < 1e-10:
                break
            diff = price - market_price
            sigma -= diff / v
            if sigma <= 0:
                sigma = 1e-6
            if abs(diff) < 1e-8:
                break
        return {"ok": True, "result": round(sigma, 8)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
