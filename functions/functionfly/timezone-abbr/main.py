from datetime import datetime
from zoneinfo import ZoneInfo

def handler(event):
    try:
        tz_name = event.get("timezone", "") if isinstance(event, dict) else ""
        date_str = event.get("date") if isinstance(event, dict) else None
        if not tz_name:
            return {"ok": False, "error": "Input timezone is required"}
        try:
            z = ZoneInfo(tz_name)
            dt = datetime.fromisoformat(date_str.replace("Z", "+00:00")) if date_str else datetime.now(z)
            abbr = dt.strftime("%Z")
            return {"ok": True, "abbreviation": abbr}
        except Exception as e:
            return {"ok": False, "error": str(e)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
