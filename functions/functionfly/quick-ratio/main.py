def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    cash = event.get("cash")
    ar = event.get("accounts_receivable")
    cl = event.get("current_liabilities")
    if cash is None or ar is None or cl is None:
        return {"ok": False, "error": "cash, accounts_receivable, and current_liabilities are required"}
    try:
        cash = float(cash)
        ar = float(ar)
        cl = float(cl)
        sti = float(event.get("short_term_investments", 0))
        if cl == 0:
            return {"ok": False, "error": "current_liabilities cannot be zero"}
        ratio = (cash + sti + ar) / cl
        return {"ok": True, "result": round(ratio, 6)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
