"""Property Tax Calculator - Calculate property tax."""


def handler(event):
    try:
        property_value = float(event.get("property_value", 0))
        tax_rate = float(event.get("tax_rate", 0))
        homestead_exemption = event.get("homestead_exemption")

        if property_value <= 0:
            return {"ok": False, "error": "property_value must be positive"}
        if tax_rate < 0:
            return {"ok": False, "error": "tax_rate cannot be negative"}
        if homestead_exemption is not None:
            homestead_exemption = float(homestead_exemption)
            if homestead_exemption < 0:
                return {"ok": False, "error": "homestead_exemption cannot be negative"}

        assessed_value = round(property_value * 0.92, 2)

        exemption = 0.0
        if homestead_exemption is not None:
            exemption = min(homestead_exemption, assessed_value)

        taxable_value = max(0, assessed_value - exemption)

        annual_tax = round(taxable_value * (tax_rate / 1000.0), 2)
        monthly_tax = round(annual_tax / 12, 2)

        return {
            "ok": True,
            "assessed_value": assessed_value,
            "exemption": exemption,
            "taxable_value": taxable_value,
            "annual_tax": annual_tax,
            "monthly_tax": monthly_tax
        }
    except (ValueError, TypeError) as e:
        return {"ok": False, "error": f"Invalid input: {str(e)}"}
    except Exception as e:
        return {"ok": False, "error": f"Internal error: {str(e)}"}
