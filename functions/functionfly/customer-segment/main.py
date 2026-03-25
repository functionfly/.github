def handler(event):
    total_spent = float(event.get("total_spent", 0)) if isinstance(event, dict) else 0
    order_count = int(event.get("order_count", 0))
    days_since_last_order = float(event.get("days_since_last_order", 9999))
    customer_age_days = float(event.get("customer_age_days", 1))
    avg_order_value = total_spent / max(order_count, 1)
    try:
        # RFM segmentation
        if days_since_last_order <= 30 and order_count >= 5 and avg_order_value >= 100:
            segment = "champions"
            description = "High-value, recent, frequent buyers"
        elif days_since_last_order <= 90 and order_count >= 3:
            segment = "loyal"
            description = "Frequent buyers with good recency"
        elif days_since_last_order <= 30 and order_count <= 2:
            segment = "new_customers"
            description = "Recently acquired customers"
        elif days_since_last_order <= 90 and avg_order_value >= 150:
            segment = "potential_loyalists"
            description = "High-value but infrequent buyers"
        elif days_since_last_order > 180 and order_count >= 3:
            segment = "at_risk"
            description = "Once loyal, now drifting away"
        elif days_since_last_order > 360:
            segment = "lost"
            description = "Likely churned customers"
        elif order_count == 0:
            segment = "prospects"
            description = "Registered but never purchased"
        else:
            segment = "occasional"
            description = "Low-frequency buyers"
        return {
            "ok": True,
            "result": segment,
            "segment": segment,
            "description": description,
            "rfm": {
                "recency_days": days_since_last_order,
                "frequency": order_count,
                "monetary": total_spent,
                "avg_order_value": round(avg_order_value, 2)
            }
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
