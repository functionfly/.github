from datetime import datetime, timedelta


def handler(event):
    if isinstance(event, dict):
        a = event.get("date_a", "")
        b = event.get("date_b", "")
    else:
        a, b = "", ""
    if not a or not b:
        return {"ok": False, "error": "Input 'date_a' and 'date_b' are required"}
    try:
        dt_a = datetime.fromisoformat(a.replace("Z", "+00:00")).date()
        dt_b = datetime.fromisoformat(b.replace("Z", "+00:00")).date()
        if dt_a > dt_b:
            dt_a, dt_b = dt_b, dt_a
        count = 0
        d = dt_a
        while d <= dt_b:
            if d.weekday() < 5:
                count += 1
            d += timedelta(days=1)
        return {"ok": True, "business_days": count}
    except Exception as e:
        return {"ok": False, "error": str(e)}
