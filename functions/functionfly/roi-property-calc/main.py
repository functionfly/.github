"""ROI Property Calculator - Calculate property investment ROI."""


def handler(event):
    try:
        purchase_price = float(event.get("purchase_price", 0))
        down_payment = float(event.get("down_payment", 0))
        closing_costs = float(event.get("closing_costs", 0))
        renovation_costs = float(event.get("renovation_costs", 0))
        monthly_rent = float(event.get("monthly_rent", 0))
        monthly_expenses = float(event.get("monthly_expenses", 0))
        annual_property_tax = float(event.get("annual_property_tax", 0))
        annual_insurance = float(event.get("annual_insurance", 0))
        holding_period_years = int(event.get("holding_period_years", 1))

        if purchase_price <= 0:
            return {"ok": False, "error": "purchase_price must be positive"}
        if down_payment < 0:
            return {"ok": False, "error": "down_payment cannot be negative"}
        if closing_costs < 0:
            return {"ok": False, "error": "closing_costs cannot be negative"}
        if renovation_costs < 0:
            return {"ok": False, "error": "renovation_costs cannot be negative"}
        if monthly_rent < 0:
            return {"ok": False, "error": "monthly_rent cannot be negative"}
        if monthly_expenses < 0:
            return {"ok": False, "error": "monthly_expenses cannot be negative"}
        if annual_property_tax < 0:
            return {"ok": False, "error": "annual_property_tax cannot be negative"}
        if annual_insurance < 0:
            return {"ok": False, "error": "annual_insurance cannot be negative"}
        if holding_period_years <= 0 or holding_period_years > 100:
            return {"ok": False, "error": "holding_period_years must be between 1 and 100"}

        total_investment = round(down_payment + closing_costs + renovation_costs, 2)

        if total_investment == 0:
            return {"ok": False, "error": "total_investment cannot be zero"}

        annual_rent = monthly_rent * 12
        annual_expenses = (monthly_expenses * 12) + annual_property_tax + annual_insurance
        net_operating_income = round(annual_rent - annual_expenses, 2)

        cash_on_cash_roi = round((net_operating_income / total_investment) * 100, 2) if total_investment > 0 else 0
        cap_rate = round((net_operating_income / purchase_price) * 100, 2) if purchase_price > 0 else 0

        return {
            "ok": True,
            "total_investment": total_investment,
            "annual_rent": round(annual_rent, 2),
            "annual_expenses": round(annual_expenses, 2),
            "net_operating_income": net_operating_income,
            "cash_on_cash_roi": cash_on_cash_roi,
            "cap_rate": cap_rate
        }
    except (ValueError, TypeError) as e:
        return {"ok": False, "error": f"Invalid input: {str(e)}"}
    except Exception as e:
        return {"ok": False, "error": f"Internal error: {str(e)}"}
