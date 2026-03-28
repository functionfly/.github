def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    nol = event.get("net_operating_loss")
    taxable_income = event.get("taxable_income")
    tax_rate = event.get("tax_rate")
    if nol is None or taxable_income is None or tax_rate is None:
        return {"ok": False, "error": "net_operating_loss, taxable_income, and tax_rate are required"}
    try:
        nol = float(nol)
        taxable_income = float(taxable_income)
        tax_rate = float(tax_rate)
        applied = min(nol, taxable_income)
        remaining = max(nol - taxable_income, 0)
        tax_benefit = applied * tax_rate
        return {
            "ok": True,
            "result": round(tax_benefit, 2),
            "applied_loss": round(applied, 2),
            "remaining_loss": round(remaining, 2),
            "tax_benefit": round(tax_benefit, 2)
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
