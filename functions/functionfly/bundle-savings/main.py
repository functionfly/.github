def handler(event):
    individual_prices = event.get("individual_prices") if isinstance(event, dict) else None
    bundle_price = event.get("bundle_price")
    if not individual_prices or bundle_price is None:
        return {"ok": False, "error": "individual_prices and bundle_price are required"}
    try:
        total_individual = round(sum(float(p) for p in individual_prices), 2)
        bundle = float(bundle_price)
        savings = round(total_individual - bundle, 2)
        savings_pct = round(savings / total_individual * 100, 2) if total_individual else 0
        return {
            "ok": True,
            "result": savings,
            "savings": savings,
            "savings_pct": savings_pct,
            "total_individual": total_individual,
            "bundle_price": bundle,
            "item_count": len(individual_prices)
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
