def handler(event):
    order_date = event.get("order_date") if isinstance(event, dict) else None
    receive_date = event.get("receive_date")
    supplier_days = event.get("supplier_days")
    processing_days = float(event.get("processing_days", 0))
    transit_days = float(event.get("transit_days", 0))
    try:
        if order_date and receive_date:
            from datetime import datetime
            fmt = "%Y-%m-%d"
            o = datetime.strptime(str(order_date), fmt)
            r = datetime.strptime(str(receive_date), fmt)
            lead_time = (r - o).days
        elif supplier_days is not None:
            lead_time = float(supplier_days) + processing_days + transit_days
        else:
            lead_time = processing_days + transit_days
        return {
            "ok": True,
            "result": lead_time,
            "lead_time_days": lead_time,
            "processing_days": processing_days,
            "transit_days": transit_days
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
