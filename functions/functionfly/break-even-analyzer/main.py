"""Break-Even Analyzer - Calculate break-even point."""
import math


def handler(event):
    try:
        fixed_costs = float(event.get("fixed_costs", 0))
        variable_cost_per_unit = float(event.get("variable_cost_per_unit", 0))
        price_per_unit = float(event.get("price_per_unit", 0))

        if fixed_costs < 0:
            return {"ok": False, "error": "fixed_costs cannot be negative"}
        if variable_cost_per_unit < 0:
            return {"ok": False, "error": "variable_cost_per_unit cannot be negative"}
        if price_per_unit <= 0:
            return {"ok": False, "error": "price_per_unit must be positive"}

        if variable_cost_per_unit >= price_per_unit:
            return {"ok": False, "error": "variable_cost_per_unit must be less than price_per_unit for break-even to be possible"}

        contribution_margin = price_per_unit - variable_cost_per_unit
        break_even_units = math.ceil(fixed_costs / contribution_margin)
        break_even_revenue = round(break_even_units * price_per_unit, 2)
        margin_of_safety_percent = 0.0

        return {
            "ok": True,
            "break_even_units": break_even_units,
            "break_even_revenue": break_even_revenue,
            "margin_of_safety_percent": margin_of_safety_percent
        }
    except (ValueError, TypeError) as e:
        return {"ok": False, "error": f"Invalid input: {str(e)}"}
    except Exception as e:
        return {"ok": False, "error": f"Internal error: {str(e)}"}
