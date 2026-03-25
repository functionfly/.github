def handler(event):
    avg_order_value = event.get("avg_order_value") if isinstance(event, dict) else None
    purchase_frequency = event.get("purchase_frequency")
    customer_lifespan_years = float(event.get("customer_lifespan_years", 3))
    profit_margin_pct = float(event.get("profit_margin_pct", 100))
    discount_rate_pct = float(event.get("discount_rate_pct", 0))
    if avg_order_value is None or purchase_frequency is None:
        return {"ok": False, "error": "avg_order_value and purchase_frequency (per year) are required"}
    try:
        aov = float(avg_order_value)
        freq = float(purchase_frequency)
        margin = profit_margin_pct / 100
        annual_revenue = aov * freq
        gross_ltv = round(annual_revenue * customer_lifespan_years * margin, 2)
        if discount_rate_pct > 0:
            r = discount_rate_pct / 100
            n = int(customer_lifespan_years)
            discounted_ltv = round(sum(annual_revenue * margin / (1 + r) ** y for y in range(1, n + 1)), 2)
        else:
            discounted_ltv = gross_ltv
        return {
            "ok": True,
            "result": gross_ltv,
            "ltv": gross_ltv,
            "discounted_ltv": discounted_ltv,
            "annual_revenue": round(annual_revenue, 2),
            "avg_order_value": aov,
            "purchase_frequency": freq,
            "lifespan_years": customer_lifespan_years
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
