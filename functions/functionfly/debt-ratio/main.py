def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    liabilities = event.get("total_liabilities")
    assets = event.get("total_assets")
    if liabilities is None or assets is None:
        return {"ok": False, "error": "total_liabilities and total_assets are required"}
    try:
        liabilities = float(liabilities)
        assets = float(assets)
        if assets == 0:
            return {"ok": False, "error": "total_assets cannot be zero"}
        ratio = liabilities / assets
        return {"ok": True, "result": round(ratio, 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
