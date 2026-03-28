import math
import datetime


def julian_day(year, month, day):
    if month <= 2:
        year -= 1
        month += 12
    A = int(year / 100)
    B = 2 - A + int(A / 4)
    return int(365.25 * (year + 4716)) + int(30.6001 * (month + 1)) + day + B - 1524.5


def moon_phase(jd):
    """Calculate moon phase for given Julian Day.
    Returns (phase_angle, illumination, age_days).
    """
    # Known new moon: January 6, 2000 at 18:14 UTC = JD 2451550.26
    known_new_moon_jd = 2451550.26
    synodic_period = 29.53058867  # days

    # Days since known new moon
    days_since = jd - known_new_moon_jd
    age_days = days_since % synodic_period
    if age_days < 0:
        age_days += synodic_period

    # Phase angle (0 = new moon, 180 = full moon)
    phase_angle = (age_days / synodic_period) * 360

    # Illumination (0 = new, 1 = full)
    illumination = (1 - math.cos(math.radians(phase_angle))) / 2 * 100

    return phase_angle, illumination, age_days


def phase_name_and_emoji(phase_angle, age_days):
    """Get phase name and emoji from phase angle."""
    synodic = 29.53058867

    if age_days < 1.85 or age_days > synodic - 1.85:
        return "New Moon", "🌑"
    elif age_days < 7.38:
        return "Waxing Crescent", "🌒"
    elif age_days < 9.22:
        return "First Quarter", "🌓"
    elif age_days < 14.77:
        return "Waxing Gibbous", "🌔"
    elif age_days < 16.61:
        return "Full Moon", "🌕"
    elif age_days < 22.15:
        return "Waning Gibbous", "🌖"
    elif age_days < 23.99:
        return "Last Quarter", "🌗"
    else:
        return "Waning Crescent", "🌘"


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}

    date_str = event.get("date")
    if date_str:
        try:
            date = datetime.date.fromisoformat(date_str)
        except ValueError:
            return {"ok": False, "error": "date must be in YYYY-MM-DD format"}
    else:
        date = datetime.date.today()

    try:
        jd = julian_day(date.year, date.month, date.day)
        phase_angle, illumination, age_days = moon_phase(jd)
        name, emoji = phase_name_and_emoji(phase_angle, age_days)

        return {
            "ok": True,
            "result": {
                "date": date.isoformat(),
                "phase_name": name,
                "phase_angle": round(phase_angle, 2),
                "illumination": round(illumination, 2),
                "age_days": round(age_days, 2),
                "emoji": emoji
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"moon phase calculation failed: {str(e)}"}
