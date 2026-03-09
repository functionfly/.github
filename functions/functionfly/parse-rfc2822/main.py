from email.utils import parsedate_to_datetime

def handler(event):
    s = event.get("date_string", "") if isinstance(event, dict) else ""
    if not s:
        return {"ok": False, "error": "Input date_string is required"}
    try:
        dt = parsedate_to_datetime(s.strip())
        return {"ok": True, "iso": dt.isoformat()}
    except Exception as e:
        return {"ok": False, "error": str(e)}
