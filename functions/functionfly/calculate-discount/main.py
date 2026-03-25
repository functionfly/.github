def handler(event):
    price = event.get("price") if isinstance(event, dict) else None
    discount = event.get("discount")
    discount_type = event.get("type", "percent")
    if price is None or discount is None:
        return {"ok": False, "error": "price and discount are required"}
    try:
        p, d = float(price), float(discount)
        if discount_type == "percent":
            if not 0 <= d <= 100:
                return {"ok": False, "error": "percent discount must be 0-100"}
            amount = round(p * d / 100, 2)
        elif discount_type == "amount":
            amount = min(round(d, 2), p)
        else:
            return {"ok": False, "error": "type must be 'percent' or 'amount'"}
        final = round(p - amount, 2)
        return {"ok": True, "result": amount, "original": p, "discount_amount": amount, "final_price": final, "savings_pct": round(amount/p*100,2) if p else 0}
    except Exception as e:
        return {"ok": False, "error": str(e)}
