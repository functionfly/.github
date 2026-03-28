def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    net_income = event.get("net_income")
    interest = event.get("interest")
    taxes = event.get("taxes")
    if net_income is None or interest is None or taxes is None:
        return {"ok": False, "error": "net_income, interest, and taxes are required"}
    try:
        ebit = float(net_income) + float(interest) + float(taxes)
        return {"ok": True, "result": round(ebit, 2)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
