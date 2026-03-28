def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    corp_yield = event.get("corporate_yield")
    rfr = event.get("risk_free_rate")
    if corp_yield is None or rfr is None:
        return {"ok": False, "error": "corporate_yield and risk_free_rate are required"}
    try:
        corp_yield = float(corp_yield)
        rfr = float(rfr)
        spread = corp_yield - rfr
        spread_bps = spread * 10000
        return {"ok": True, "result": round(spread, 8), "result_bps": round(spread_bps, 4)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
