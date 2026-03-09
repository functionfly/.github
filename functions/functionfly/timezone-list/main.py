from zoneinfo import available_timezones

def handler(event):
    f = (event.get("filter") or "").lower() if isinstance(event, dict) else ""
    zones = sorted(available_timezones())
    if f:
        zones = [z for z in zones if f in z.lower()]
    return {"ok": True, "timezones": zones}
