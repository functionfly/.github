from datetime import datetime


def handler(event):
    if isinstance(event, dict):
        date_str = event.get("date", "")
    else:
        date_str = ""
    if not date_str:
        return {"ok": False, "error": "Input 'date' is required"}
    try:
        dt = datetime.fromisoformat(date_str.replace("Z", "+00:00"))
        quarter = (dt.month - 1) // 3 + 1
        return {"ok": True, "quarter": quarter}
    except Exception as e:
        return {"ok": False, "error": str(e)}
