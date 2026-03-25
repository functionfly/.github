def handler(event):
    returns = event.get("returns") if isinstance(event, dict) else None
    total_orders = event.get("total_orders")
    period = event.get("period", "monthly")
    reasons = event.get("return_reasons", [])
    if returns is None or total_orders is None:
        return {"ok": False, "error": "returns and total_orders are required"}
    try:
        r, t = int(returns), int(total_orders)
        if t == 0:
            return {"ok": False, "error": "total_orders must be > 0"}
        rate = round(r / t * 100, 4)
        benchmarks = {"fashion": 30, "electronics": 12, "books": 5, "home": 15, "food": 2}
        from collections import Counter
        reason_dist = Counter(str(reason) for reason in reasons)
        return {
            "ok": True,
            "result": rate,
            "return_rate_pct": rate,
            "returns": r,
            "total_orders": t,
            "period": period,
            "industry_benchmarks": benchmarks,
            "top_reasons": reason_dist.most_common(5)
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
