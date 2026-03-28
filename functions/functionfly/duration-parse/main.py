import re

def parse_iso8601_duration(duration: str) -> float:
    """Parse ISO 8601 duration string to seconds"""
    pattern = r'^P(?:(\d+)Y)?(?:(\d+)M)?(?:(\d+)D)?(?:T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+(?:\.\d+)?)S)?)?$'
    match = re.match(pattern, duration)
    if not match:
        raise ValueError(f"Invalid ISO 8601 duration: {duration}")
    years = int(match.group(1) or 0)
    months = int(match.group(2) or 0)
    days = int(match.group(3) or 0)
    hours = int(match.group(4) or 0)
    minutes = int(match.group(5) or 0)
    seconds = float(match.group(6) or 0)
    total_seconds = years * 365.25 * 24 * 3600 + months * 30.44 * 24 * 3600 + days * 24 * 3600 + hours * 3600 + minutes * 60 + seconds
    return total_seconds

def parse_human_duration(duration: str) -> float:
    """Parse human-readable duration string to seconds"""
    duration = duration.lower().strip()
    total_seconds = 0
    patterns = [
        (r'(\d+(?:\.\d+)?)\s*(?:years?|y)', 365.25 * 24 * 3600),
        (r'(\d+(?:\.\d+)?)\s*(?:months?|mo)', 30.44 * 24 * 3600),
        (r'(\d+(?:\.\d+)?)\s*(?:weeks?|w)', 7 * 24 * 3600),
        (r'(\d+(?:\.\d+)?)\s*(?:days?|d)', 24 * 3600),
        (r'(\d+(?:\.\d+)?)\s*(?:hours?|h)', 3600),
        (r'(\d+(?:\.\d+)?)\s*(?:minutes?|mins?|m)', 60),
        (r'(\d+(?:\.\d+)?)\s*(?:seconds?|secs?|s)', 1),
    ]
    for pattern, multiplier in patterns:
        matches = re.findall(pattern, duration)
        for match in matches:
            total_seconds += float(match) * multiplier
    if total_seconds == 0:
        raise ValueError(f"Invalid duration format: {duration}")
    return total_seconds

def handler(event):
    try:
        duration = event.get("duration", "") if isinstance(event, dict) else ""
        if not duration:
            return {"ok": False, "error": "duration is required"}
        try:
            if duration.startswith('P'):
                seconds = parse_iso8601_duration(duration)
            else:
                seconds = parse_human_duration(duration)
            return {"ok": True, "seconds": seconds, "minutes": seconds / 60, "hours": seconds / 3600, "days": seconds / 86400}
        except ValueError as e:
            return {"ok": False, "error": str(e)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
