def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    fixed = event.get("fixed_costs")
    var_cost = event.get("variable_cost_per_unit")
    price = event.get("price_per_unit")
    if fixed is None or var_cost is None or price is None:
        return {"ok": False, "error": "fixed_costs, variable_cost_per_unit, and price_per_unit are required"}
    try:
        fixed = float(fixed)
        var_cost = float(var_cost)
        price = float(price)
        contribution_margin = price - var_cost
        if contribution_margin <= 0:
            return {"ok": False, "error": "price_per_unit must be greater than variable_cost_per_unit"}
        units = fixed / contribution_margin
        revenue = units * price
        return {
            "ok": True,
            "result": round(units, 4),
            "break_even_units": round(units, 4),
            "break_even_revenue": round(revenue, 2),
            "contribution_margin": round(contribution_margin, 2)
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
