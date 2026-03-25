def handler(event):
    days_since_last_purchase = event.get("days_since_last_purchase") if isinstance(event, dict) else None
    avg_purchase_interval_days = float(event.get("avg_purchase_interval_days", 30))
    total_orders = int(event.get("total_orders", 1))
    customer_age_days = float(event.get("customer_age_days", 365))
    if days_since_last_purchase is None:
        return {"ok": False, "error": "days_since_last_purchase is required"}
    try:
        idle = float(days_since_last_purchase)
        # Recency score (0-1, lower is better)
        recency_score = min(1.0, idle / (avg_purchase_interval_days * 3))
        # Frequency score
        expected_orders = customer_age_days / max(1, avg_purchase_interval_days)
        freq_score = 1 - min(1.0, total_orders / max(1, expected_orders))
        # Weighted churn probability
        churn_probability = round((recency_score * 0.7 + freq_score * 0.3), 4)
        if churn_probability < 0.3:
            risk = "low"
        elif churn_probability < 0.6:
            risk = "medium"
        elif churn_probability < 0.8:
            risk = "high"
        else:
            risk = "critical"
        return {
            "ok": True,
            "result": churn_probability,
            "churn_probability": churn_probability,
            "churn_risk": risk,
            "days_since_purchase": idle,
            "recency_score": round(recency_score, 4),
            "frequency_score": round(freq_score, 4)
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
