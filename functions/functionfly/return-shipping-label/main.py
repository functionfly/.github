import os, hashlib, time


def handler(event):
    order_id = event.get("order_id") if isinstance(event, dict) else None
    from_address = event.get("from_address", {})
    to_address = event.get("to_address", {})
    weight_kg = float(event.get("weight_kg", 0.5))
    carrier = event.get("carrier", "standard")
    if not order_id:
        return {"ok": False, "error": "order_id is required"}
    try:
        ts = int(time.time())
        tracking_seed = f"{order_id}{ts}{os.urandom(4).hex()}"
        tracking_num = "RETURN" + hashlib.md5(tracking_seed.encode()).hexdigest()[:12].upper()
        label = {
            "tracking_number": tracking_num,
            "carrier": carrier.upper(),
            "from": from_address,
            "to": to_address,
            "weight_kg": weight_kg,
            "created_at": ts,
            "order_id": str(order_id),
            "label_type": "return",
            "estimated_cost": round(3.99 + weight_kg * 1.5, 2),
            "instructions": f"Print this label and drop off at any {carrier.upper()} location."
        }
        return {"ok": True, "result": tracking_num, "label": label, "tracking_number": tracking_num}
    except Exception as e:
        return {"ok": False, "error": str(e)}
