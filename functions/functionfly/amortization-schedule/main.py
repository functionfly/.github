def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    principal = event.get("principal")
    annual_rate = event.get("annual_rate")
    years = event.get("years")
    if principal is None or annual_rate is None or years is None:
        return {"ok": False, "error": "principal, annual_rate, and years are required"}
    try:
        principal = float(principal)
        annual_rate = float(annual_rate)
        years = float(years)
        periods_per_year = int(event.get("periods_per_year", 12))
        n = int(years * periods_per_year)
        r = annual_rate / periods_per_year
        if r == 0:
            payment = principal / n
        else:
            payment = principal * r * (1 + r) ** n / ((1 + r) ** n - 1)
        schedule = []
        balance = principal
        for period in range(1, n + 1):
            interest = balance * r
            principal_paid = payment - interest
            balance -= principal_paid
            if abs(balance) < 0.01:
                balance = 0
            schedule.append({
                "period": period,
                "payment": round(payment, 2),
                "principal": round(principal_paid, 2),
                "interest": round(interest, 2),
                "balance": round(max(balance, 0), 2)
            })
        return {"ok": True, "result": schedule, "payment": round(payment, 2), "total_periods": n}
    except Exception as e:
        return {"ok": False, "error": str(e)}
