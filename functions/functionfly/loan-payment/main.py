def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    principal = event.get("principal")
    annual_rate = event.get("annual_rate")
    years = event.get("years")
    if principal is None or annual_rate is None or years is None:
        return {"ok": False, "error": "principal, annual_rate, and years are required"}
    try:
        principal = float(principal)
        annual_rate = float(annual_rate)
        years = float(years)
        n = years * 12
        r = annual_rate / 12
        if r == 0:
            monthly = principal / n
        else:
            monthly = principal * r * (1 + r) ** n / ((1 + r) ** n - 1)
        total = monthly * n
        total_interest = total - principal
        return {
            "ok": True,
            "result": round(monthly, 2),
            "monthly_payment": round(monthly, 2),
            "total_payment": round(total, 2),
            "total_interest": round(total_interest, 2)
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
