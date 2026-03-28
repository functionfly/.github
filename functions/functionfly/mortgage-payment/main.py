def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    home_price = event.get("home_price")
    down_payment = event.get("down_payment")
    annual_rate = event.get("annual_rate")
    years = event.get("years")
    if home_price is None or down_payment is None or annual_rate is None or years is None:
        return {"ok": False, "error": "home_price, down_payment, annual_rate, and years are required"}
    try:
        home_price = float(home_price)
        down_payment = float(down_payment)
        annual_rate = float(annual_rate)
        years = float(years)
        loan_amount = home_price - down_payment
        if loan_amount <= 0:
            return {"ok": False, "error": "down_payment must be less than home_price"}
        n = years * 12
        r = annual_rate / 12
        if r == 0:
            monthly = loan_amount / n
        else:
            monthly = loan_amount * r * (1 + r) ** n / ((1 + r) ** n - 1)
        total = monthly * n
        return {
            "ok": True,
            "result": round(monthly, 2),
            "monthly_payment": round(monthly, 2),
            "loan_amount": round(loan_amount, 2),
            "total_payment": round(total, 2),
            "total_interest": round(total - loan_amount, 2)
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
