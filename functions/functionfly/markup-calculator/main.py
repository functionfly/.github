"""Markup Calculator - Calculate markup from cost."""


def handler(event):
    try:
        cost = float(event.get("cost", 0))
        markup_percent = float(event.get("markup_percent", 0))

        if cost < 0:
            return {"ok": False, "error": "cost cannot be negative"}
        if markup_percent < 0:
            return {"ok": False, "error": "markup_percent cannot be negative"}

        selling_price = round(cost * (1 + markup_percent / 100.0), 2)
        gross_profit = round(selling_price - cost, 2)
        gross_margin_percent = round((gross_profit / selling_price) * 100.0, 2) if selling_price > 0 else 0

        return {
            "ok": True,
            "selling_price": selling_price,
            "gross_profit": gross_profit,
            "gross_margin_percent": gross_margin_percent
        }
    except (ValueError, TypeError) as e:
        return {"ok": False, "error": f"Invalid input: {str(e)}"}
    except Exception as e:
        return {"ok": False, "error": f"Internal error: {str(e)}"}
