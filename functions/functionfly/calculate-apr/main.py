def handler(event):
    principal = event.get("principal") if isinstance(event, dict) else None
    fees = float(event.get("fees", 0))
    monthly_payment = event.get("monthly_payment")
    months = int(event.get("months", 12))
    if principal is None:
        return {"ok": False, "error": "principal is required"}
    try:
        p = float(principal)
        if monthly_payment:
            pmt = float(monthly_payment)
            total_paid = pmt * months
            total_interest = total_paid - p + fees
            apr = round((total_interest / (p + fees)) / months * 12 * 100, 4)
        else:
            apr = round(fees / p * 12 / months * 100, 4)
        return {"ok": True, "result": apr, "apr_pct": apr, "principal": p, "fees": fees, "months": months}
    except Exception as e:
        return {"ok": False, "error": str(e)}
