def handler(event):
    weight = event.get("weight") if isinstance(event, dict) else None
    distance = event.get("distance")
    method = event.get("method", "standard")
    base_rate = float(event.get("base_rate", 5.0))
    per_kg = float(event.get("per_kg", 1.5))
    if weight is None:
        return {"ok": False, "error": "weight is required (kg)"}
    try:
        w = float(weight)
        multipliers = {"standard": 1.0, "express": 2.0, "overnight": 3.5, "economy": 0.7}
        mult = multipliers.get(method, 1.0)
        cost = round((base_rate + w * per_kg) * mult, 2)
        if distance:
            d = float(distance)
            cost = round(cost + d * 0.01 * mult, 2)
        return {"ok": True, "result": cost, "weight_kg": w, "method": method, "shipping_cost": cost}
    except Exception as e:
        return {"ok": False, "error": str(e)}
