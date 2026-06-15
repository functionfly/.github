"""Inventory Turnover Calculator - Calculate inventory turnover ratio."""


def handler(event):
    try:
        cogs = float(event.get("cogs", 0))
        beginning_inventory = float(event.get("beginning_inventory", 0))
        ending_inventory = float(event.get("ending_inventory", 0))

        if cogs < 0:
            return {"ok": False, "error": "cogs cannot be negative"}
        if beginning_inventory < 0:
            return {"ok": False, "error": "beginning_inventory cannot be negative"}
        if ending_inventory < 0:
            return {"ok": False, "error": "ending_inventory cannot be negative"}

        avg_inventory = (beginning_inventory + ending_inventory) / 2.0

        if avg_inventory == 0:
            return {"ok": False, "error": "average inventory cannot be zero"}

        turnover_ratio = round(cogs / avg_inventory, 2)
        days_in_inventory = round(365.0 / turnover_ratio, 2) if turnover_ratio > 0 else 365

        return {
            "ok": True,
            "turnover_ratio": turnover_ratio,
            "days_in_inventory": days_in_inventory,
            "avg_inventory": round(avg_inventory, 2)
        }
    except (ValueError, TypeError) as e:
        return {"ok": False, "error": f"Invalid input: {str(e)}"}
    except Exception as e:
        return {"ok": False, "error": f"Internal error: {str(e)}"}
