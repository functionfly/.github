from collections import Counter


def handler(event):
    purchases = event.get("purchases") if isinstance(event, dict) else None
    if not purchases:
        return {"ok": False, "error": "purchases is required (list of orders)"}
    try:
        total_spent = sum(float(p.get("amount", 0)) for p in purchases)
        total_orders = len(purchases)
        avg_order = round(total_spent / total_orders, 2) if total_orders else 0
        categories = Counter(str(p.get("category", "unknown")) for p in purchases)
        favorite_category = categories.most_common(1)[0][0] if categories else None
        amounts = sorted([float(p.get("amount", 0)) for p in purchases])
        n = len(amounts)
        median = amounts[n // 2] if n % 2 else round((amounts[n // 2 - 1] + amounts[n // 2]) / 2, 2)
        max_order = max(amounts) if amounts else 0
        min_order = min(amounts) if amounts else 0
        if n >= 2 and all(p.get("date") for p in purchases):
            from datetime import datetime
            dates = sorted(datetime.fromisoformat(str(p["date"])) for p in purchases)
            intervals = [(dates[i+1]-dates[i]).days for i in range(len(dates)-1)]
            avg_interval = round(sum(intervals)/len(intervals), 1) if intervals else None
        else:
            avg_interval = None
        return {
            "ok": True,
            "result": total_spent,
            "total_spent": round(total_spent, 2),
            "total_orders": total_orders,
            "avg_order_value": avg_order,
            "median_order_value": median,
            "max_order": max_order,
            "min_order": min_order,
            "favorite_category": favorite_category,
            "category_breakdown": dict(categories),
            "avg_purchase_interval_days": avg_interval
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
