def handler(event):
    try:
        seconds = event.get("seconds") if isinstance(event, dict) else None
        minutes = event.get("minutes") if isinstance(event, dict) else None
        hours = event.get("hours") if isinstance(event, dict) else None
        days = event.get("days") if isinstance(event, dict) else None
        total_seconds = 0
        if seconds is not None:
            total_seconds = seconds
        elif minutes is not None:
            total_seconds = minutes * 60
        elif hours is not None:
            total_seconds = hours * 3600
        elif days is not None:
            total_seconds = days * 86400
        else:
            return {"ok": False, "error": "at least one duration value is required"}
        if total_seconds < 0:
            return {"ok": False, "error": "duration must be positive"}
        days_val = int(total_seconds // 86400)
        hours_val = int((total_seconds % 86400) // 3600)
        minutes_val = int((total_seconds % 3600) // 60)
        seconds_val = int(total_seconds % 60)
        parts = []
        if days_val > 0:
            parts.append(f"{days_val} day{'s' if days_val != 1 else ''}")
        if hours_val > 0:
            parts.append(f"{hours_val} hour{'s' if hours_val != 1 else ''}")
        if minutes_val > 0:
            parts.append(f"{minutes_val} minute{'s' if minutes_val != 1 else ''}")
        if seconds_val > 0:
            parts.append(f"{seconds_val} second{'s' if seconds_val != 1 else ''}")
        if not parts:
            human = "0 seconds"
        else:
            human = " ".join(parts)
        return {"ok": True, "human": human}
    except Exception as e:
        return {"ok": False, "error": str(e)}
