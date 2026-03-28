import math


# Timezone database: list of (min_lat, max_lat, min_lng, max_lng, timezone, utc_offset, abbr, country)
TIMEZONE_DB = [
    # North America
    (-90, 90, -180, -168, "Pacific/Honolulu", -10, "HST", "United States"),
    (54, 72, -168, -141, "America/Anchorage", -9, "AKST", "United States"),
    (18, 72, -141, -120, "America/Anchorage", -9, "AKST", "United States"),
    (18, 72, -120, -105, "America/Los_Angeles", -8, "PST", "United States"),
    (18, 72, -105, -90, "America/Denver", -7, "MST", "United States"),
    (18, 72, -90, -75, "America/Chicago", -6, "CST", "United States"),
    (18, 72, -75, -60, "America/New_York", -5, "EST", "United States"),
    (18, 72, -60, -52, "America/Halifax", -4, "AST", "Canada"),
    (18, 72, -52, -45, "America/St_Johns", -3.5, "NST", "Canada"),
    # South America
    (-60, 18, -82, -60, "America/Bogota", -5, "COT", "Colombia"),
    (-60, 18, -60, -45, "America/Sao_Paulo", -3, "BRT", "Brazil"),
    (-60, 18, -45, -30, "America/Sao_Paulo", -3, "BRT", "Brazil"),
    (-60, 18, -82, -75, "America/Lima", -5, "PET", "Peru"),
    (-60, 18, -75, -65, "America/La_Paz", -4, "BOT", "Bolivia"),
    (-60, 18, -65, -55, "America/Buenos_Aires", -3, "ART", "Argentina"),
    # Europe
    (35, 72, -10, 0, "Europe/London", 0, "GMT", "United Kingdom"),
    (35, 72, 0, 15, "Europe/Paris", 1, "CET", "France"),
    (35, 72, 15, 30, "Europe/Helsinki", 2, "EET", "Finland"),
    (35, 72, 30, 45, "Europe/Moscow", 3, "MSK", "Russia"),
    # Africa
    (-35, 35, -20, 15, "Africa/Lagos", 1, "WAT", "Nigeria"),
    (-35, 35, 15, 35, "Africa/Nairobi", 3, "EAT", "Kenya"),
    (-35, 35, 35, 55, "Africa/Nairobi", 3, "EAT", "Kenya"),
    # Middle East
    (15, 45, 35, 50, "Asia/Riyadh", 3, "AST", "Saudi Arabia"),
    (15, 45, 50, 65, "Asia/Dubai", 4, "GST", "UAE"),
    (15, 45, 65, 80, "Asia/Karachi", 5, "PKT", "Pakistan"),
    # Asia
    (5, 55, 65, 80, "Asia/Kolkata", 5.5, "IST", "India"),
    (5, 55, 80, 90, "Asia/Dhaka", 6, "BST", "Bangladesh"),
    (5, 55, 90, 105, "Asia/Bangkok", 7, "ICT", "Thailand"),
    (5, 55, 105, 120, "Asia/Singapore", 8, "SGT", "Singapore"),
    (5, 55, 120, 135, "Asia/Shanghai", 8, "CST", "China"),
    (5, 55, 135, 145, "Asia/Tokyo", 9, "JST", "Japan"),
    (5, 55, 145, 160, "Asia/Seoul", 9, "KST", "South Korea"),
    # Australia/Pacific
    (-45, 5, 110, 130, "Australia/Perth", 8, "AWST", "Australia"),
    (-45, 5, 130, 145, "Australia/Darwin", 9.5, "ACST", "Australia"),
    (-45, 5, 145, 155, "Australia/Sydney", 10, "AEST", "Australia"),
    (-45, 5, 155, 180, "Pacific/Auckland", 12, "NZST", "New Zealand"),
    # Russia
    (45, 90, 45, 60, "Asia/Yekaterinburg", 5, "YEKT", "Russia"),
    (45, 90, 60, 75, "Asia/Omsk", 6, "OMST", "Russia"),
    (45, 90, 75, 90, "Asia/Krasnoyarsk", 7, "KRAT", "Russia"),
    (45, 90, 90, 105, "Asia/Irkutsk", 8, "IRKT", "Russia"),
    (45, 90, 105, 120, "Asia/Yakutsk", 9, "YAKT", "Russia"),
    (45, 90, 120, 135, "Asia/Vladivostok", 10, "VLAT", "Russia"),
    (45, 90, 135, 180, "Asia/Magadan", 11, "MAGT", "Russia"),
]


def find_timezone(lat, lng):
    """Find timezone for given coordinates."""
    best = None
    best_area = float("inf")

    for entry in TIMEZONE_DB:
        min_lat, max_lat, min_lng, max_lng = entry[0], entry[1], entry[2], entry[3]
        if min_lat <= lat <= max_lat and min_lng <= lng <= max_lng:
            area = (max_lat - min_lat) * (max_lng - min_lng)
            if area < best_area:
                best_area = area
                best = entry

    if best:
        return best[4], best[5], best[6], best[7]

    # Fallback: estimate from longitude
    utc_offset = round(lng / 15)
    utc_offset = max(-12, min(14, utc_offset))
    sign = "+" if utc_offset >= 0 else "-"
    tz_name = f"Etc/GMT{sign}{abs(utc_offset)}" if utc_offset != 0 else "UTC"
    return tz_name, utc_offset, "UTC", "Unknown"


def format_offset(offset):
    hours = int(offset)
    minutes = int(abs(offset - hours) * 60)
    sign = "+" if offset >= 0 else "-"
    if minutes:
        return f"UTC{sign}{abs(hours)}:{minutes:02d}"
    return f"UTC{sign}{abs(hours)}:00"


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

    try:
        timezone, utc_offset, abbreviation, country = find_timezone(lat, lng)

        return {
            "ok": True,
            "result": {
                "timezone": timezone,
                "utc_offset": utc_offset,
                "utc_offset_str": format_offset(utc_offset),
                "abbreviation": abbreviation,
                "country": country
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"timezone lookup failed: {str(e)}"}
