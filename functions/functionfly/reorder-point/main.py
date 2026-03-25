import math


def handler(event):
    daily_demand = event.get("daily_demand") if isinstance(event, dict) else None
    lead_time_days = event.get("lead_time_days")
    safety_stock_days = float(event.get("safety_stock_days", 2))
    demand_std_dev = event.get("demand_std_dev")
    service_level_z = float(event.get("service_level_z", 1.65))  # 95% service level
    if daily_demand is None or lead_time_days is None:
        return {"ok": False, "error": "daily_demand and lead_time_days are required"}
    try:
        d, lt = float(daily_demand), float(lead_time_days)
        demand_during_lead_time = d * lt
        if demand_std_dev is not None:
            safety_stock = round(service_level_z * float(demand_std_dev) * math.sqrt(lt))
        else:
            safety_stock = round(d * safety_stock_days)
        rop = round(demand_during_lead_time + safety_stock)
        return {
            "ok": True,
            "result": rop,
            "reorder_point": rop,
            "safety_stock": safety_stock,
            "demand_during_lead_time": round(demand_during_lead_time, 2),
            "daily_demand": d,
            "lead_time_days": lt
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
