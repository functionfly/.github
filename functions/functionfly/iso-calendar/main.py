from datetime import datetime

def handler(event):
    try:
        date_str = event.get("date") if isinstance(event, dict) else None
        if date_str:
            try:
                dt = datetime.fromisoformat(date_str.replace("Z", "+00:00"))
            except ValueError:
                return {"ok": False, "error": "invalid date format"}
        else:
            dt = datetime.now()
        iso_cal = dt.isocalendar()
        return {"ok": True, "year": iso_cal[0], "week": iso_cal[1], "weekday": iso_cal[2]}
    except Exception as e:
        return {"ok": False, "error": str(e)}
