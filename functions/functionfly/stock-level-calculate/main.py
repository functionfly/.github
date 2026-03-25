def handler(event):
    opening_stock = event.get("opening_stock") if isinstance(event, dict) else None
    units_received = int(event.get("units_received", 0))
    units_sold = int(event.get("units_sold", 0))
    units_returned = int(event.get("units_returned", 0))
    units_damaged = int(event.get("units_damaged", 0))
    units_reserved = int(event.get("units_reserved", 0))
    if opening_stock is None:
        return {"ok": False, "error": "opening_stock is required"}
    try:
        opening = int(opening_stock)
        closing = opening + units_received - units_sold + units_returned - units_damaged
        available = max(0, closing - units_reserved)
        turnover = round(units_sold / max(1, (opening + closing) / 2), 4) if units_sold else 0
        return {
            "ok": True,
            "result": closing,
            "opening_stock": opening,
            "closing_stock": closing,
            "available_stock": available,
            "units_sold": units_sold,
            "units_received": units_received,
            "inventory_turnover": turnover
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
