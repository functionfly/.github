import math
import datetime


def julian_day(year, month, day):
    """Calculate Julian Day Number."""
    if month <= 2:
        year -= 1
        month += 12
    A = int(year / 100)
    B = 2 - A + int(A / 4)
    return int(365.25 * (year + 4716)) + int(30.6001 * (month + 1)) + day + B - 1524.5


def sun_position(jd):
    """Calculate sun position for given Julian Day."""
    n = jd - 2451545.0  # Days since J2000.0
    L = (280.460 + 0.9856474 * n) % 360  # Mean longitude
    g = math.radians((357.528 + 0.9856003 * n) % 360)  # Mean anomaly
    lam = math.radians(L + 1.915 * math.sin(g) + 0.020 * math.sin(2 * g))  # Ecliptic longitude
    eps = math.radians(23.439 - 0.0000004 * n)  # Obliquity

    # Right ascension and declination
    sin_lam = math.sin(lam)
    cos_lam = math.cos(lam)
    sin_eps = math.sin(eps)
    cos_eps = math.cos(eps)

    ra = math.atan2(cos_eps * sin_lam, cos_lam)
    dec = math.asin(sin_eps * sin_lam)

    return ra, dec


def equation_of_time(jd):
    """Calculate equation of time in minutes."""
    n = jd - 2451545.0
    L = math.radians((280.460 + 0.9856474 * n) % 360)
    g = math.radians((357.528 + 0.9856003 * n) % 360)
    eps = math.radians(23.439 - 0.0000004 * n)

    E = -2 * math.sin(g) + 4 * math.sin(g) * math.cos(2 * L)
    E -= 0.5 * math.sin(2 * L) * math.cos(2 * L)
    E -= 1.25 * math.sin(2 * g)
    return E * (180 / math.pi) * 4  # Convert to minutes


def sunrise_sunset(lat, lng, jd, zenith_deg=90.833):
    """Calculate sunrise and sunset times.
    Returns (sunrise_utc_hours, sunset_utc_hours) or None if polar day/night.
    """
    lat_r = math.radians(lat)
    _, dec = sun_position(jd)
    zenith_r = math.radians(zenith_deg)

    cos_hour_angle = (math.cos(zenith_r) - math.sin(lat_r) * math.sin(dec)) / (math.cos(lat_r) * math.cos(dec))

    if cos_hour_angle < -1:
        return None, None, True, False  # Polar day
    if cos_hour_angle > 1:
        return None, None, False, True  # Polar night

    hour_angle = math.degrees(math.acos(cos_hour_angle))

    eot = equation_of_time(jd)
    solar_noon_utc = 12 - lng / 15 - eot / 60

    sunrise_utc = solar_noon_utc - hour_angle / 15
    sunset_utc = solar_noon_utc + hour_angle / 15

    return sunrise_utc, sunset_utc, False, False


def hours_to_hhmm(hours):
    """Convert decimal hours to HH:MM string."""
    hours = hours % 24
    h = int(hours)
    m = int((hours - h) * 60)
    return f"{h:02d}:{m:02d}"


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}

    lat = event.get("lat")
    lng = event.get("lng")

    if lat is None:
        return {"ok": False, "error": "lat is required"}
    if lng is None:
        return {"ok": False, "error": "lng is required"}

    try:
        lat = float(lat)
        lng = float(lng)
    except (TypeError, ValueError):
        return {"ok": False, "error": "lat and lng must be numbers"}

    if not (-90 <= lat <= 90):
        return {"ok": False, "error": "lat must be between -90 and 90"}
    if not (-180 <= lng <= 180):
        return {"ok": False, "error": "lng must be between -180 and 180"}

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

        sunrise, sunset, polar_day, polar_night = sunrise_sunset(lat, lng, jd)
        civil_begin, civil_end, _, _ = sunrise_sunset(lat, lng, jd, zenith_deg=96.0)

        result = {
            "date": date.isoformat(),
            "polar_day": polar_day,
            "polar_night": polar_night
        }

        if polar_day:
            result["note"] = "Sun does not set on this date at this location"
        elif polar_night:
            result["note"] = "Sun does not rise on this date at this location"
        else:
            result["sunrise"] = hours_to_hhmm(sunrise)
            result["sunset"] = hours_to_hhmm(sunset)

            # Solar noon
            eot = equation_of_time(jd)
            solar_noon = 12 - lng / 15 - eot / 60
            result["solar_noon"] = hours_to_hhmm(solar_noon)

            day_length = sunset - sunrise
            result["day_length_hours"] = round(day_length, 2)

            if civil_begin is not None:
                result["civil_twilight_begin"] = hours_to_hhmm(civil_begin)
            if civil_end is not None:
                result["civil_twilight_end"] = hours_to_hhmm(civil_end)

        return {"ok": True, "result": result}

    except Exception as e:
        return {"ok": False, "error": f"sun times calculation failed: {str(e)}"}
