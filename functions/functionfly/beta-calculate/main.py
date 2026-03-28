def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    asset_returns = event.get("asset_returns")
    market_returns = event.get("market_returns")
    if asset_returns is None or market_returns is None:
        return {"ok": False, "error": "asset_returns and market_returns are required"}
    try:
        a = [float(x) for x in asset_returns]
        m = [float(x) for x in market_returns]
        if len(a) != len(m):
            return {"ok": False, "error": "asset_returns and market_returns must have same length"}
        if len(a) < 2:
            return {"ok": False, "error": "At least 2 data points required"}
        n = len(a)
        mean_a = sum(a) / n
        mean_m = sum(m) / n
        cov = sum((a[i] - mean_a) * (m[i] - mean_m) for i in range(n)) / (n - 1)
        var_m = sum((m[i] - mean_m) ** 2 for i in range(n)) / (n - 1)
        if var_m == 0:
            return {"ok": False, "error": "Market returns have zero variance"}
        beta = cov / var_m
        return {"ok": True, "result": round(beta, 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
