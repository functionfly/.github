def handler(event):
    current_price = event.get("current_price") if isinstance(event, dict) else None
    alert_price = event.get("alert_price")
    alert_type = event.get("alert_type", "below")
    original_price = event.get("original_price")
    if current_price is None or alert_price is None:
        return {"ok": False, "error": "current_price and alert_price are required"}
    try:
        cur = float(current_price)
        alert = float(alert_price)
        triggered = (alert_type == "below" and cur <= alert) or \
                    (alert_type == "above" and cur >= alert) or \
                    (alert_type == "percent_drop" and original_price is not None and
                     (float(original_price) - cur) / float(original_price) * 100 >= alert)
        price_drop = round(float(original_price) - cur, 2) if original_price else None
        pct_drop = round(price_drop / float(original_price) * 100, 2) if original_price and price_drop else None
        return {
            "ok": True,
            "result": triggered,
            "triggered": triggered,
            "current_price": cur,
            "alert_price": alert,
            "alert_type": alert_type,
            "price_drop": price_drop,
            "price_drop_pct": pct_drop
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
