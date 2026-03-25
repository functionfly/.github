def handler(event):
    historical_demand = event.get("historical_demand") if isinstance(event, dict) else None
    periods_ahead = int(event.get("periods_ahead", 1))
    method = event.get("method", "moving_average")
    window = int(event.get("window", 3))
    alpha = float(event.get("alpha", 0.3))
    if not historical_demand:
        return {"ok": False, "error": "historical_demand is required (list of numbers)"}
    try:
        data = [float(x) for x in historical_demand]
        n = len(data)
        if n < 2:
            return {"ok": False, "error": "Need at least 2 historical data points"}
        if method == "moving_average":
            w = min(window, n)
            last_ma = sum(data[-w:]) / w
            forecasts = [round(last_ma, 2)] * periods_ahead
        elif method == "exponential_smoothing":
            forecast = data[0]
            for val in data[1:]:
                forecast = alpha * val + (1 - alpha) * forecast
            forecasts = [round(forecast, 2)] * periods_ahead
        elif method == "linear_trend":
            x_mean = (n - 1) / 2
            y_mean = sum(data) / n
            numerator = sum((i - x_mean) * (data[i] - y_mean) for i in range(n))
            denominator = sum((i - x_mean) ** 2 for i in range(n))
            slope = numerator / denominator if denominator else 0
            intercept = y_mean - slope * x_mean
            forecasts = [round(intercept + slope * (n + i), 2) for i in range(periods_ahead)]
        else:
            return {"ok": False, "error": "method must be 'moving_average', 'exponential_smoothing', or 'linear_trend'"}
        avg = round(sum(data) / n, 2)
        std = round((sum((x - avg) ** 2 for x in data) / n) ** 0.5, 2)
        return {
            "ok": True,
            "result": forecasts[0] if periods_ahead == 1 else forecasts,
            "forecasts": forecasts,
            "method": method,
            "historical_mean": avg,
            "historical_std": std
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
