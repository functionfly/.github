import math

# Z-score lookup for common confidence levels
Z_SCORES = {0.90: 1.2816, 0.95: 1.6449, 0.99: 2.3263, 0.995: 2.5758, 0.999: 3.0902}


def _z_score(confidence):
    if confidence in Z_SCORES:
        return Z_SCORES[confidence]
    # Approximate inverse normal using Abramowitz & Stegun
    if confidence < 0.5:
        confidence = 1 - confidence
    t = math.sqrt(-2 * math.log(1 - confidence))
    c0, c1, c2 = 2.515517, 0.802853, 0.010328
    d1, d2, d3 = 1.432788, 0.189269, 0.001308
    return t - (c0 + c1 * t + c2 * t * t) / (1 + d1 * t + d2 * t * t + d3 * t * t * t)


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    portfolio_value = event.get("portfolio_value")
    volatility = event.get("volatility")
    if portfolio_value is None:
        return {"ok": False, "error": "portfolio_value is required"}
    if volatility is None:
        return {"ok": False, "error": "volatility is required"}
    try:
        pv = float(portfolio_value)
        vol = float(volatility)
        confidence = float(event.get("confidence_level", 0.99))
        horizon = float(event.get("time_horizon", 1))
        if not (0 < confidence < 1):
            return {"ok": False, "error": "confidence_level must be between 0 and 1"}
        if vol < 0:
            return {"ok": False, "error": "volatility must be non-negative"}
        z = _z_score(confidence)
        economic_capital = pv * vol * math.sqrt(horizon) * z
        risk_weighted_pct = (economic_capital / pv * 100) if pv > 0 else 0
        return {
            "ok": True,
            "result": round(economic_capital, 2),
            "economic_capital": round(economic_capital, 2),
            "portfolio_value": pv,
            "volatility": vol,
            "confidence_level": confidence,
            "time_horizon": horizon,
            "z_score": round(z, 4),
            "risk_weighted_pct": round(risk_weighted_pct, 2),
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
