import math
import datetime


CONDITIONS = [
    ("Clear", 0),
    ("Partly Cloudy", 0),
    ("Cloudy", 2),
    ("Light Rain", 5),
    ("Rain", 15),
    ("Heavy Rain", 30),
    ("Thunderstorm", 25),
    ("Snow", 8),
    ("Sleet", 5),
    ("Fog", 0),
    ("Windy", 0),
    ("Haze", 0)
]


def base_temp(lat):
    if abs(lat) < 10:
        return 28
    elif abs(lat) < 23.5:
        return 25
    elif abs(lat) < 35:
        return 20
    elif abs(lat) < 50:
        return 12
    elif abs(lat) < 60:
        return 5
    elif abs(lat) < 70:
        return -5
    else:
        return -20


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

    days = event.get("days", 5)
    try:
        days = max(1, min(7, int(days)))
    except (TypeError, ValueError):
        days = 5

    units = event.get("units", "metric")
    if units not in ("metric", "imperial"):
        return {"ok": False, "error": "units must be 'metric' or 'imperial'"}

    try:
        today = datetime.date.today()
        h_base = int(abs(lat * 100) + abs(lng * 100)) % 1000

        forecast = []
        for day_offset in range(days):
            forecast_date = today + datetime.timedelta(days=day_offset)
            day_of_year = forecast_date.timetuple().tm_yday

            # Seasonal factor
            season_factor = math.cos(2 * math.pi * (day_of_year - 172) / 365)
            if lat < 0:
                season_factor = -season_factor

            bt = base_temp(lat)
            temp_mid = bt + season_factor * 15

            # Day-to-day variation
            h = (h_base + day_offset * 37) % 100
            temp_variation = (h - 50) / 5

            temp_high_c = round(temp_mid + 5 + temp_variation, 1)
            temp_low_c = round(temp_mid - 5 + temp_variation, 1)

            # Conditions
            cond_idx = (h_base + day_offset * 13) % len(CONDITIONS)
            # Adjust for temperature
            if temp_high_c < 0 and cond_idx in (3, 4, 5):
                cond_idx = 7  # Snow
            conditions, precip_base = CONDITIONS[cond_idx]
            precipitation_mm = round(precip_base + (h % 10) * 0.5, 1) if precip_base > 0 else 0

            # Humidity
            humidity = min(100, max(20, 60 + (h % 40) - 20))

            # Wind
            wind_speed = round(5 + (h % 25), 1)

            if units == "imperial":
                temp_high = round(temp_high_c * 9/5 + 32, 1)
                temp_low = round(temp_low_c * 9/5 + 32, 1)
                wind = round(wind_speed * 0.621371, 1)
                precip = round(precipitation_mm / 25.4, 2)
                temp_unit = "°F"
                wind_unit = "mph"
                precip_unit = "in"
            else:
                temp_high = temp_high_c
                temp_low = temp_low_c
                wind = wind_speed
                precip = precipitation_mm
                temp_unit = "°C"
                wind_unit = "km/h"
                precip_unit = "mm"

            forecast.append({
                "date": forecast_date.isoformat(),
                "temp_high": temp_high,
                "temp_low": temp_low,
                "conditions": conditions,
                "precipitation": precip,
                "humidity": humidity,
                "wind_speed": wind,
                "temp_unit": temp_unit,
                "wind_unit": wind_unit,
                "precip_unit": precip_unit
            })

        return {
            "ok": True,
            "result": {
                "location": {"lat": lat, "lng": lng},
                "forecast": forecast,
                "units": units,
                "note": "Simulated forecast data - use OpenWeatherMap or similar for real forecasts"
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"weather forecast failed: {str(e)}"}
