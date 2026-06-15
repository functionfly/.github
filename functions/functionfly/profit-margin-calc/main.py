"""Profit Margin Calculator - Calculate profit margins."""


def handler(event):
    try:
        revenue = float(event.get("revenue", 0))
        cogs = float(event.get("cogs", 0))
        operating_expenses = event.get("operating_expenses")

        if revenue <= 0:
            return {"ok": False, "error": "revenue must be positive"}
        if cogs < 0:
            return {"ok": False, "error": "cogs cannot be negative"}
        if operating_expenses is not None:
            operating_expenses = float(operating_expenses)
            if operating_expenses < 0:
                return {"ok": False, "error": "operating_expenses cannot be negative"}

        gross_profit = revenue - cogs
        gross_margin_percent = round((gross_profit / revenue) * 100, 2) if revenue > 0 else 0

        result = {
            "ok": True,
            "gross_margin": round(gross_profit, 2),
            "gross_margin_percent": gross_margin_percent,
            "net_income": round(gross_profit, 2)
        }

        if operating_expenses is not None:
            operating_margin = gross_profit - operating_expenses
            operating_margin_percent = round((operating_margin / revenue) * 100, 2) if revenue > 0 else 0
            net_income = operating_margin
            net_margin_percent = round((net_income / revenue) * 100, 2) if revenue > 0 else 0

            result["operating_margin"] = round(operating_margin, 2)
            result["operating_margin_percent"] = operating_margin_percent
            result["net_margin_percent"] = net_margin_percent
            result["net_income"] = round(net_income, 2)

        return result
    except (ValueError, TypeError) as e:
        return {"ok": False, "error": f"Invalid input: {str(e)}"}
    except Exception as e:
        return {"ok": False, "error": f"Internal error: {str(e)}"}
