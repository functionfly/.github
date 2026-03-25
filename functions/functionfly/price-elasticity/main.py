def handler(event):
    price_change_pct = event.get("price_change_pct") if isinstance(event, dict) else None
    demand_change_pct = event.get("demand_change_pct")
    price_a = event.get("price_a")
    demand_a = event.get("demand_a")
    price_b = event.get("price_b")
    demand_b = event.get("demand_b")
    try:
        if price_change_pct is not None and demand_change_pct is not None:
            pct_p = float(price_change_pct)
            pct_d = float(demand_change_pct)
        elif all(x is not None for x in [price_a, demand_a, price_b, demand_b]):
            pa, da, pb, db = float(price_a), float(demand_a), float(price_b), float(demand_b)
            pct_p = (pb - pa) / pa * 100 if pa else 0
            pct_d = (db - da) / da * 100 if da else 0
        else:
            return {"ok": False, "error": "Provide price_change_pct+demand_change_pct or price_a+demand_a+price_b+demand_b"}
        if pct_p == 0:
            return {"ok": False, "error": "price_change_pct cannot be 0"}
        elasticity = round(pct_d / pct_p, 4)
        if abs(elasticity) > 1:
            classification = "elastic"
            interpretation = "Demand is sensitive to price changes"
        elif abs(elasticity) < 1:
            classification = "inelastic"
            interpretation = "Demand is insensitive to price changes"
        else:
            classification = "unit_elastic"
            interpretation = "Demand changes proportionally to price"
        return {
            "ok": True,
            "result": elasticity,
            "elasticity": elasticity,
            "abs_elasticity": abs(elasticity),
            "classification": classification,
            "interpretation": interpretation,
            "price_change_pct": pct_p,
            "demand_change_pct": pct_d
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
