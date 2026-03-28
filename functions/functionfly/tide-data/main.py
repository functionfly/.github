import math
import datetime


def simulate_tides(lat, lng, date, days):
    """Generate simulated tide data based on location and date."""
    # Use a simple sinusoidal model with location-based variation
    # Real tides depend on lunar/solar cycles and local geography

    # Seed based on location for consistent results
    seed = int(abs(lat * 100) + abs(lng * 100)) % 1000

    # Base tidal range varies by latitude (simplified)
    if abs(lat) < 10:
        tidal_range = 0.5  # Tropics - small tides
    elif abs(lat) < 30:
        tidal_range = 1.2
    elif abs(lat) < 50:
        tidal_range = 1.8
    else:
        tidal_range = 2.5  # Higher latitudes - larger tides

    # Mean water level
    mean_level = tidal_range / 2

    tides = []
    for day_offset in range(days):
        current_date = date + datetime.timedelta(days=day_offset)

        # Day of year for seasonal variation
        day_of_year = current_date.timetuple().tm_yday

        # Lunar cycle approximation (29.5 day period)
        lunar_phase = (day_of_year + seed) % 29.5 / 29.5

        # Generate 4 tides per day (semi-diurnal pattern)
        # High tides at approximately 6-hour intervals
        base_hour = (seed % 6) + lunar_phase * 6

        for i in range(4):
            hour = (base_hour + i * 6.2) % 24
            is_high = i % 2 == 0

            # Height variation
            if is_high:
                height = mean_level + tidal_range * (0.8 + 0.2 * math.sin(lunar_phase * 2 * math.pi))
            else:
                height = mean_level - tidal_range * (0.7 + 0.1 * math.sin(lunar_phase * 2 * math.pi))

            height = max(0.05, round(height, 2))

            h = int(hour)
            m = int((hour - h) * 60)

            tides.append({
                "date": current_date.isoformat(),
                "time": f"{h:02d}:{m:02d}",
                "height_m": height,
                "type": "high" if is_high else "low"
            })

    return sorted(tides, key=lambda x: (x["date"], x["time"]))


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

    days = event.get("days", 1)
    try:
        days = int(days)
        days = max(1, min(7, days))
    except (TypeError, ValueError):
        days = 1

    try:
        tides = simulate_tides(lat, lng, date, days)

        return {
            "ok": True,
            "result": {
                "location": {"lat": lat, "lng": lng},
                "date": date.isoformat(),
                "days": days,
                "tides": tides,
                "unit": "meters",
                "note": "Simulated tide data - use NOAA or similar service for accurate predictions"
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"tide data generation failed: {str(e)}"}
