from datetime import datetime

def date_to_julian_day(dt: datetime) -> float:
    """Convert datetime to Julian Day Number"""
    year = dt.year
    month = dt.month
    day = dt.day
    hour = dt.hour
    minute = dt.minute
    second = dt.second
    microsecond = dt.microsecond
    if month <= 2:
        year -= 1
        month += 12
    A = year // 100
    B = 2 - A + A // 4
    JD = int(365.25 * (year + 4716)) + int(30.6001 * (month + 1)) + day + B - 1524.5
    JD += (hour + minute / 60 + second / 3600 + microsecond / 3600000000) / 24
    return JD

def julian_day_to_date(jd: float) -> datetime:
    """Convert Julian Day Number to datetime"""
    JD = jd + 0.5
    Z = int(JD)
    F = JD - Z
    if Z < 2299161:
        A = Z
    else:
        alpha = int((Z - 1867216.25) / 36524.25)
        A = Z + 1 + alpha - alpha // 4
    B = A + 1524
    C = int((B - 122.1) / 365.25)
    D = int(365.25 * C)
    E = int((B - D) / 30.6001)
    day = B - D - int(30.6001 * E) + F
    if E < 14:
        month = E - 1
    else:
        month = E - 13
    if month > 2:
        year = C - 4716
    else:
        year = C - 4715
    day_int = int(day)
    day_frac = day - day_int
    hours = int(day_frac * 24)
    minutes = int((day_frac * 24 - hours) * 60)
    seconds = int(((day_frac * 24 - hours) * 60 - minutes) * 60)
    return datetime(year, month, day_int, hours, minutes, seconds)

def handler(event):
    try:
        date_str = event.get("date") if isinstance(event, dict) else None
        julian_day = event.get("julian_day") if isinstance(event, dict) else None
        if date_str:
            try:
                dt = datetime.fromisoformat(date_str.replace("Z", "+00:00"))
                jd = date_to_julian_day(dt)
                return {"ok": True, "julian_day": jd}
            except ValueError:
                return {"ok": False, "error": "invalid date format"}
        elif julian_day is not None:
            try:
                dt = julian_day_to_date(julian_day)
                return {"ok": True, "date": dt.isoformat() + "Z"}
            except Exception as e:
                return {"ok": False, "error": f"conversion error: {str(e)}"}
        else:
            dt = datetime.now()
            jd = date_to_julian_day(dt)
            return {"ok": True, "julian_day": jd}
    except Exception as e:
        return {"ok": False, "error": str(e)}
