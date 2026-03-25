import re


def _parse_iso8601_duration(s):
    m = re.match(r'PT(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?', s)
    if not m:
        return None
    h = int(m.group(1) or 0)
    mn = int(m.group(2) or 0)
    sc = int(m.group(3) or 0)
    return h * 3600 + mn * 60 + sc


def handler(event):
    duration = event.get("duration") if isinstance(event, dict) else None
    if not duration:
        return {"ok": False, "error": "duration is required (ISO 8601 PT#H#M#S or seconds)"}
    try:
        d = str(duration).strip()
        if d.upper().startswith("PT"):
            total_seconds = _parse_iso8601_duration(d.upper())
            if total_seconds is None:
                return {"ok": False, "error": f"Invalid ISO 8601 duration: {d}"}
        else:
            total_seconds = int(float(d))
        h = total_seconds // 3600
        m = (total_seconds % 3600) // 60
        s = total_seconds % 60
        formatted = f"{h}:{m:02d}:{s:02d}" if h else f"{m}:{s:02d}"
        return {
            "ok": True,
            "result": total_seconds,
            "total_seconds": total_seconds,
            "hours": h,
            "minutes": m,
            "seconds": s,
            "formatted": formatted,
            "iso8601": f"PT{h}H{m}M{s}S" if h else f"PT{m}M{s}S"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
