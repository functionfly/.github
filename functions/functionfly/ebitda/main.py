def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    required = ["net_income", "interest", "taxes", "depreciation", "amortization"]
    for field in required:
        if event.get(field) is None:
            return {"ok": False, "error": f"{field} is required"}
    try:
        net_income = float(event["net_income"])
        interest = float(event["interest"])
        taxes = float(event["taxes"])
        depreciation = float(event["depreciation"])
        amortization = float(event["amortization"])
        ebitda = net_income + interest + taxes + depreciation + amortization
        return {"ok": True, "result": round(ebitda, 2)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
