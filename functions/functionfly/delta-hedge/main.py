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
        num_contracts = float(event.get("num_contracts", 1))
        contract_size = float(event.get("contract_size", 100))
        if T <= 0:
            return {"ok": False, "error": "time_to_expiry must be positive"}
        d1 = (math.log(S / K) + (r + 0.5 * sigma ** 2) * T) / (sigma * math.sqrt(T))
        if option_type == "call":
            delta = _norm_cdf(d1)
        else:
            delta = _norm_cdf(d1) - 1
        shares_to_hedge = delta * num_contracts * contract_size
        return {
            "ok": True,
            "result": {
                "delta": round(delta, 6),
                "shares_to_short": round(shares_to_hedge, 4),
                "hedge_cost": round(shares_to_hedge * S, 2)
            }
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
