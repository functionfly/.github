import math


def handler(event):
    base_price = event.get("base_price") if isinstance(event, dict) else None
    demand_level = float(event.get("demand_level", 0.5))
    stock_level = float(event.get("stock_level", 0.5))
    time_of_day_factor = float(event.get("time_of_day_factor", 1.0))
    competitor_price = event.get("competitor_price")
    min_price = float(event.get("min_price", 0))
    max_price_multiplier = float(event.get("max_price_multiplier", 2.0))
    if base_price is None:
        return {"ok": False, "error": "base_price is required"}
    try:
        base = float(base_price)
        # Demand factor: higher demand = higher price
        demand_factor = 0.8 + (demand_level * 0.6)
        # Supply factor: lower stock = higher price
        supply_factor = 1.0 + max(0, (0.2 - stock_level)) * 1.5
        # Competitor-based adjustment
        if competitor_price is not None:
            comp = float(competitor_price)
            comp_factor = 0.95 + (base / comp * 0.1) if comp > 0 else 1.0
        else:
            comp_factor = 1.0
        raw_price = base * demand_factor * supply_factor * time_of_day_factor * comp_factor
        dynamic_price = round(max(min_price, min(raw_price, base * max_price_multiplier)), 2)
        return {
            "ok": True,
            "result": dynamic_price,
            "dynamic_price": dynamic_price,
            "base_price": base,
            "multiplier": round(dynamic_price / base, 4),
            "factors": {"demand": round(demand_factor, 4), "supply": round(supply_factor, 4), "time_of_day": time_of_day_factor}
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
