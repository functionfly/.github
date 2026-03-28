import math
import datetime


def simulate_weather(lat, lng, units):
    """Generate simulated weather based on location and current season."""
    now = datetime.datetime.utcnow()
    day_of_year = now.timetuple().tm_yday
    hour = now.hour

    # Seasonal temperature variation
    # Northern hemisphere: summer in June-August, winter in Dec-Feb
    # Southern hemisphere: opposite
    season_factor = math.cos(2 * math.pi * (day_of_year - 172) / 365)
    if lat < 0:
        season_factor = -season_factor  # Flip for southern hemisphere

    # Base temperature by latitude (tropical = warm, polar = cold)
    if abs(lat) < 10:
        base_temp = 28  # Tropical
    elif abs(lat) < 23.5:
        base_temp = 25  # Subtropical
    elif abs(lat) < 35:
        base_temp = 20  # Warm temperate
    elif abs(lat) < 50:
        base_temp = 12  # Temperate
    elif abs(lat) < 60:
        base_temp = 5   # Cool temperate
    elif abs(lat) < 70:
        base_temp = -5  # Subarctic
    else:
        base_temp = -20  # Polar

    # Apply seasonal variation
    temp_c = base_temp + season_factor * 15

    # Diurnal variation (warmer in afternoon)
    diurnal = math.sin(math.pi * (hour - 6) / 12) * 5
    temp_c += diurnal

    # Add some variation based on location hash
    h = int(abs(lat * 100) + abs(lng * 100)) % 100
    temp_c += (h - 50) / 10

    temp_c = round(temp_c, 1)

    # Humidity (higher near equator and coasts)
    base_humidity = 60
    if abs(lat) < 10:
        base_humidity = 80
    elif abs(lat) > 60:
        base_humidity = 70
    humidity = min(100, max(10, base_humidity + (h % 30) - 15))

    # Wind
    wind_speed = round(5 + (h % 20), 1)
    wind_dir = (h * 37) % 360

    # Pressure
    pressure = round(1013.25 + (h - 50) * 0.5, 2)

    # Conditions
    conditions_list = [
        ("Clear", "clear sky"),
        ("Partly Cloudy", "partly cloudy"),
        ("Cloudy", "overcast clouds"),
        ("Light Rain", "light rain"),
        ("Rain", "moderate rain"),
        ("Thunderstorm", "thunderstorm"),
        ("Snow", "light snow"),
        ("Fog", "foggy conditions"),
        ("Windy", "strong winds"),
        ("Haze", "hazy conditions")
    ]

    cond_idx = h % len(conditions_list)
    # Adjust for temperature
    if temp_c < -5 and cond_idx in (3, 4):
        cond_idx = 6  # Snow instead of rain
    elif temp_c > 30 and cond_idx == 7:
        cond_idx = 9  # Haze instead of fog

    conditions, description = conditions_list[cond_idx]

    # Visibility
    if conditions in ("Fog",):
        visibility = round(0.5 + (h % 5) * 0.2, 1)
    elif conditions in ("Rain", "Thunderstorm"):
        visibility = round(3 + (h % 5), 1)
    else:
        visibility = round(8 + (h % 12), 1)

    # Feels like (wind chill / heat index)
    if temp_c < 10:
        feels_like = round(temp_c - wind_speed * 0.3, 1)
    elif temp_c > 25:
        feels_like = round(temp_c + humidity * 0.1, 1)
    else:
        feels_like = temp_c

    if units == "imperial":
        temp_f = round(temp_c * 9/5 + 32, 1)
        feels_like_f = round(feels_like * 9/5 + 32, 1)
        wind_mph = round(wind_speed * 0.621371, 1)
        return {
            "temperature": temp_f,
            "feels_like": feels_like_f,
            "humidity": humidity,
            "pressure_hpa": pressure,
            "wind_speed": wind_mph,
            "wind_direction": wind_dir,
            "conditions": conditions,
            "description": description,
            "visibility_km": visibility,
            "units": "imperial",
            "temp_unit": "°F",
            "wind_unit": "mph"
        }
    else:
        return {
            "temperature": temp_c,
            "feels_like": feels_like,
            "humidity": humidity,
            "pressure_hpa": pressure,
            "wind_speed": wind_speed,
            "wind_direction": wind_dir,
            "conditions": conditions,
            "description": description,
            "visibility_km": visibility,
            "units": "metric",
            "temp_unit": "°C",
            "wind_unit": "km/h"
        }


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

    units = event.get("units", "metric")
    if units not in ("metric", "imperial"):
        return {"ok": False, "error": "units must be 'metric' or 'imperial'"}

    try:
        weather = simulate_weather(lat, lng, units)
        weather["lat"] = lat
        weather["lng"] = lng
        weather["note"] = "Simulated weather data - use OpenWeatherMap or similar for real data"

        return {"ok": True, "result": weather}

    except Exception as e:
        return {"ok": False, "error": f"weather query failed: {str(e)}"}
