import math


def handler(event):
    current_stock = event.get("current_stock") if isinstance(event, dict) else None
    daily_demand = event.get("daily_demand")
    lead_time_days = float(event.get("lead_time_days", 7))
    reorder_quantity = event.get("reorder_quantity")
    forecast_days = int(event.get("forecast_days", 30))
    demand_growth_pct = float(event.get("demand_growth_pct", 0))
    if current_stock is None or daily_demand is None:
        return {"ok": False, "error": "current_stock and daily_demand are required"}
    try:
        stock, demand = float(current_stock), float(daily_demand)
        daily_demand_adjusted = demand * (1 + demand_growth_pct / 100 / 365)
        days_of_stock = round(stock / daily_demand_adjusted) if daily_demand_adjusted else float("inf")
        stockout_day = min(days_of_stock, forecast_days)
        reorder_point = math.ceil(daily_demand_adjusted * lead_time_days)
        reorder_day = max(0, math.floor((stock - reorder_point) / daily_demand_adjusted)) if daily_demand_adjusted else forecast_days
        will_stockout = days_of_stock < forecast_days
        timeline = []
        for day in range(0, forecast_days + 1, 7):
            proj = max(0, stock - daily_demand_adjusted * day)
            timeline.append({"day": day, "projected_stock": round(proj, 0)})
        return {
            "ok": True,
            "result": days_of_stock,
            "days_of_stock_remaining": days_of_stock,
            "will_stockout_in_forecast": will_stockout,
            "stockout_day": stockout_day if will_stockout else None,
            "reorder_point": reorder_point,
            "suggested_reorder_day": reorder_day,
            "daily_demand": round(daily_demand_adjusted, 2),
            "forecast_timeline": timeline
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
