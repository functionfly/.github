def handler(event):
    action = event.get("action", "earn") if isinstance(event, dict) else "earn"
    amount = event.get("amount")
    current_points = float(event.get("current_points", 0))
    points_rate = float(event.get("points_per_dollar", 1))
    redemption_rate = float(event.get("cents_per_point", 1))
    redeem_points = event.get("redeem_points")
    try:
        if action == "earn":
            if amount is None:
                return {"ok": False, "error": "amount required for earn action"}
            earned = int(float(amount) * points_rate)
            new_balance = int(current_points + earned)
            return {"ok": True, "result": earned, "points_earned": earned, "new_balance": new_balance}
        elif action == "redeem":
            if redeem_points is None:
                return {"ok": False, "error": "redeem_points required for redeem action"}
            rpts = int(redeem_points)
            if rpts > current_points:
                return {"ok": True, "result": 0, "redeemed": 0, "error_detail": "insufficient points", "available_points": int(current_points)}
            discount = round(rpts * redemption_rate / 100, 2)
            new_balance = int(current_points - rpts)
            return {"ok": True, "result": discount, "discount_value": discount, "points_redeemed": rpts, "new_balance": new_balance}
        elif action == "balance":
            redemption_value = round(current_points * redemption_rate / 100, 2)
            return {"ok": True, "result": int(current_points), "points": int(current_points), "redemption_value": redemption_value}
        else:
            return {"ok": False, "error": "action must be 'earn', 'redeem', or 'balance'"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
