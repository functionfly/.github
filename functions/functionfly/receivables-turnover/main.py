def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    sales = event.get("net_credit_sales")
    avg_ar = event.get("average_accounts_receivable")
    if sales is None or avg_ar is None:
        return {"ok": False, "error": "net_credit_sales and average_accounts_receivable are required"}
    try:
        sales = float(sales)
        avg_ar = float(avg_ar)
        if avg_ar == 0:
            return {"ok": False, "error": "average_accounts_receivable cannot be zero"}
        turnover = sales / avg_ar
        dso = 365 / turnover
        return {"ok": True, "result": round(turnover, 6), "days_sales_outstanding": round(dso, 2)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
