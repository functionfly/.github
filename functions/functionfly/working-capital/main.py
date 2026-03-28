def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    current_assets = event.get("current_assets")
    current_liabilities = event.get("current_liabilities")
    if current_assets is None or current_liabilities is None:
        return {"ok": False, "error": "current_assets and current_liabilities are required"}
    try:
        wc = float(current_assets) - float(current_liabilities)
        return {"ok": True, "result": round(wc, 2)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
