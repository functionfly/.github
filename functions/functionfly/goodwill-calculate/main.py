def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    purchase_price = event.get("purchase_price")
    fv_assets = event.get("fair_value_assets")
    fv_liabilities = event.get("fair_value_liabilities")
    if purchase_price is None or fv_assets is None or fv_liabilities is None:
        return {"ok": False, "error": "purchase_price, fair_value_assets, and fair_value_liabilities are required"}
    try:
        purchase_price = float(purchase_price)
        fv_assets = float(fv_assets)
        fv_liabilities = float(fv_liabilities)
        net_assets = fv_assets - fv_liabilities
        goodwill = purchase_price - net_assets
        return {"ok": True, "result": round(goodwill, 2), "net_assets": round(net_assets, 2)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
