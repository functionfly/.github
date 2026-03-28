import datetime


AQI_CATEGORIES = [
    (0, 50, "Good", "#00E400", "Air quality is satisfactory, and air pollution poses little or no risk."),
    (51, 100, "Moderate", "#FFFF00", "Air quality is acceptable. However, there may be a risk for some people, particularly those who are unusually sensitive to air pollution."),
    (101, 150, "Unhealthy for Sensitive Groups", "#FF7E00", "Members of sensitive groups may experience health effects. The general public is less likely to be affected."),
    (151, 200, "Unhealthy", "#FF0000", "Some members of the general public may experience health effects; members of sensitive groups may experience more serious health effects."),
    (201, 300, "Very Unhealthy", "#8F3F97", "Health alert: The risk of health effects is increased for everyone."),
    (301, 500, "Hazardous", "#7E0023", "Health warning of emergency conditions: everyone is more likely to be affected.")
]


def get_aqi_category(aqi):
    for low, high, cat, color, msg in AQI_CATEGORIES:
        if low <= aqi <= high:
            return cat, color, msg
    return "Hazardous", "#7E0023", "Health warning of emergency conditions."


def simulate_air_quality(lat, lng):
    """Simulate air quality based on location."""
    import math

    # Urban areas tend to have higher pollution
    # Use location hash for deterministic results
    h = int(abs(lat * 100) + abs(lng * 100)) % 1000

    # Base AQI by region type
    # Major industrial/urban areas
    high_pollution_regions = [
        # China industrial belt
        (30, 45, 100, 125),
        # India
        (20, 35, 68, 90),
        # Southeast Asia
        (-5, 25, 95, 115),
        # Middle East
        (20, 35, 35, 60),
    ]

    base_aqi = 30 + (h % 70)  # Default: 30-100

    for min_lat, max_lat, min_lng, max_lng in high_pollution_regions:
        if min_lat <= lat <= max_lat and min_lng <= lng <= max_lng:
            base_aqi = 80 + (h % 120)
            break

    # Rural/remote areas have better air quality
    if abs(lat) > 60 or (abs(lat) < 5 and (lng < -60 or lng > 100)):
        base_aqi = 10 + (h % 30)

    aqi = min(500, max(0, base_aqi))

    # Generate pollutant levels based on AQI
    pm25 = round(aqi * 0.2 + (h % 20) * 0.1, 1)
    pm10 = round(pm25 * 1.8 + (h % 10), 1)
    o3 = round(30 + (h % 60), 1)
    no2 = round(aqi * 0.15 + (h % 15), 1)
    so2 = round(aqi * 0.05 + (h % 5), 1)
    co = round(0.2 + (h % 10) * 0.05, 2)

    # Determine dominant pollutant
    pollutants = {"pm25": pm25, "pm10": pm10, "o3": o3, "no2": no2, "so2": so2, "co": co}
    dominant = max(pollutants, key=lambda k: pollutants[k] / {"pm25": 35, "pm10": 150, "o3": 70, "no2": 100, "so2": 75, "co": 9}.get(k, 1))

    return aqi, pollutants, dominant


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
        aqi, pollutants, dominant = simulate_air_quality(lat, lng)
        category, color, health_message = get_aqi_category(aqi)

        return {
            "ok": True,
            "result": {
                "lat": lat,
                "lng": lng,
                "aqi": aqi,
                "category": category,
                "color": color,
                "health_message": health_message,
                "pollutants": {
                    "pm25_ug_m3": pollutants["pm25"],
                    "pm10_ug_m3": pollutants["pm10"],
                    "o3_ppb": pollutants["o3"],
                    "no2_ppb": pollutants["no2"],
                    "so2_ppb": pollutants["so2"],
                    "co_ppm": pollutants["co"]
                },
                "dominant_pollutant": dominant,
                "note": "Simulated air quality data - use AirVisual or similar for real data"
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"air quality query failed: {str(e)}"}
