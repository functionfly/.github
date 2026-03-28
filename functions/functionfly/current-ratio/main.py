def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    current_assets = event.get("current_assets")
    current_liabilities = event.get("current_liabilities")
    if current_assets is None or current_liabilities is None:
        return {"ok": False, "error": "current_assets and current_liabilities are required"}
    try:
        current_assets = float(current_assets)
        current_liabilities = float(current_liabilities)
        if current_liabilities == 0:
            return {"ok": False, "error": "current_liabilities cannot be zero"}
        ratio = current_assets / current_liabilities
        return {"ok": True, "result": round(ratio, 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
