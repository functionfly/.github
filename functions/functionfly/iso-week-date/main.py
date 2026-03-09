from datetime import datetime

def handler(event):
    date_str = event.get("date", "") if isinstance(event, dict) else ""
    if not date_str:
        return {"ok": False, "error": "Input date is required"}
    try:
        dt = datetime.fromisoformat(date_str.replace("Z", "+00:00"))
        y, w, wd = dt.isocalendar()
        return {"ok": True, "iso_year": y, "iso_week": w, "iso_weekday": wd}
    except Exception as e:
        return {"ok": False, "error": str(e)}
