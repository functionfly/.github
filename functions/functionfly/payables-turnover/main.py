def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    cogs = event.get("cost_of_goods_sold")
    avg_ap = event.get("average_accounts_payable")
    if cogs is None or avg_ap is None:
        return {"ok": False, "error": "cost_of_goods_sold and average_accounts_payable are required"}
    try:
        cogs = float(cogs)
        avg_ap = float(avg_ap)
        if avg_ap == 0:
            return {"ok": False, "error": "average_accounts_payable cannot be zero"}
        turnover = cogs / avg_ap
        dpo = 365 / turnover
        return {"ok": True, "result": round(turnover, 6), "days_payable_outstanding": round(dpo, 2)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
