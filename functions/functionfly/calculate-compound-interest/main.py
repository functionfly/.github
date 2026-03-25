def handler(event):
    principal = event.get("principal") if isinstance(event, dict) else None
    rate = event.get("rate")
    time = float(event.get("time", 1))
    n = int(event.get("compounds_per_year", 12))
    if principal is None or rate is None:
        return {"ok": False, "error": "principal and rate are required"}
    try:
        p, r = float(principal), float(rate) / 100
        amount = round(p * (1 + r / n) ** (n * time), 2)
        interest = round(amount - p, 2)
        return {"ok": True, "result": amount, "final_amount": amount, "interest_earned": interest, "principal": p, "rate_pct": rate, "years": time}
    except Exception as e:
        return {"ok": False, "error": str(e)}
