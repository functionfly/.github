"""Mortgage Calculator - Calculate mortgage payment with taxes and insurance."""


def handler(event):
    try:
        principal = float(event.get("principal", 0))
        annual_rate = float(event.get("annual_rate", 0))
        term_years = int(event.get("term_years", 30))

        if principal <= 0:
            return {"ok": False, "error": "principal must be positive"}
        if annual_rate < 0 or annual_rate > 30:
            return {"ok": False, "error": "annual_rate must be between 0 and 30"}
        if term_years <= 0 or term_years > 50:
            return {"ok": False, "error": "term_years must be between 1 and 50"}

        monthly_rate = (annual_rate / 100.0) / 12.0
        num_payments = term_years * 12

        if monthly_rate > 0:
            monthly_payment = principal * (monthly_rate * (1 + monthly_rate)**num_payments) / ((1 + monthly_rate)**num_payments - 1)
        else:
            monthly_payment = principal / num_payments
        monthly_payment = round(monthly_payment, 2)

        total_interest = round(monthly_payment * num_payments - principal, 2)
        total_cost = round(monthly_payment * num_payments, 2)

        annual_property_tax = principal * 0.012
        monthly_property_tax = round(annual_property_tax / 12, 2)

        annual_insurance = principal * 0.005
        monthly_insurance = round(annual_insurance / 12, 2)

        total_monthly_payment = round(monthly_payment + monthly_property_tax + monthly_insurance, 2)

        return {
            "ok": True,
            "monthly_payment": monthly_payment,
            "total_interest": total_interest,
            "total_cost": total_cost,
            "monthly_property_tax": monthly_property_tax,
            "monthly_insurance": monthly_insurance,
            "total_monthly_payment": total_monthly_payment
        }
    except (ValueError, TypeError) as e:
        return {"ok": False, "error": f"Invalid input: {str(e)}"}
    except Exception as e:
        return {"ok": False, "error": f"Internal error: {str(e)}"}
