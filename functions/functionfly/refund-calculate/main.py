def handler(event):
    original_amount = event.get("original_amount") if isinstance(event, dict) else None
    refund_type = event.get("refund_type", "full")
    quantity_returned = event.get("quantity_returned")
    quantity_ordered = event.get("quantity_ordered", 1)
    restocking_fee_pct = float(event.get("restocking_fee_pct", 0))
    already_refunded = float(event.get("already_refunded", 0))
    if original_amount is None:
        return {"ok": False, "error": "original_amount is required"}
    try:
        amt = float(original_amount)
        if refund_type == "full":
            gross_refund = amt
        elif refund_type == "partial" and quantity_returned is not None:
            qty_r, qty_o = int(quantity_returned), int(quantity_ordered)
            if qty_o <= 0:
                return {"ok": False, "error": "quantity_ordered must be > 0"}
            gross_refund = round(amt * qty_r / qty_o, 2)
        elif refund_type == "custom":
            custom_amount = event.get("custom_amount")
            if custom_amount is None:
                return {"ok": False, "error": "custom_amount required for custom refund"}
            gross_refund = min(float(custom_amount), amt)
        else:
            return {"ok": False, "error": "refund_type must be full, partial, or custom"}
        restocking_fee = round(gross_refund * restocking_fee_pct / 100, 2)
        net_refund = round(max(0, gross_refund - restocking_fee - already_refunded), 2)
        return {
            "ok": True,
            "result": net_refund,
            "gross_refund": gross_refund,
            "restocking_fee": restocking_fee,
            "already_refunded": already_refunded,
            "net_refund": net_refund,
            "refund_type": refund_type
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
