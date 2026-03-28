import math

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    prices = event.get("prices")
    if prices is None:
        return {"ok": False, "error": "prices is required"}
    try:
        prices = [float(p) for p in prices]
        if len(prices) < 2:
            return {"ok": False, "error": "At least 2 prices required"}
        log_returns = [math.log(prices[i] / prices[i - 1]) for i in range(1, len(prices))]
        n = len(log_returns)
        mean = sum(log_returns) / n
        variance = sum((r - mean) ** 2 for r in log_returns) / (n - 1)
        daily_vol = math.sqrt(variance)
        periods_per_year = float(event.get("periods_per_year", 252))
        annual_vol = daily_vol * math.sqrt(periods_per_year)
        return {
            "ok": True,
            "result": round(annual_vol, 8),
            "daily_volatility": round(daily_vol, 8),
            "annualized_volatility": round(annual_vol, 8)
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
