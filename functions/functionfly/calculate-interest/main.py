def handler(event):
    principal = event.get("principal") if isinstance(event, dict) else None
    rate = event.get("rate")
    time = event.get("time", 1)
    time_unit = event.get("time_unit", "years")
    if principal is None or rate is None:
        return {"ok": False, "error": "principal and rate are required"}
    try:
        p, r, t = float(principal), float(rate), float(time)
        if time_unit == "months":
            t = t / 12
        elif time_unit == "days":
            t = t / 365
        interest = round(p * r / 100 * t, 2)
        return {"ok": True, "result": interest, "interest": interest, "total": round(p + interest, 2), "principal": p}
    except Exception as e:
        return {"ok": False, "error": str(e)}
