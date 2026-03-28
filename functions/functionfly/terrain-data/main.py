import math


def classify_terrain(elevation_m, slope_deg):
    """Classify terrain type based on elevation and slope."""
    if elevation_m < 0:
        return "below_sea_level", "flat"
    elif elevation_m < 50 and slope_deg < 2:
        return "coastal_plain", "flat"
    elif elevation_m < 200 and slope_deg < 5:
        return "lowland", "gentle"
    elif elevation_m < 500 and slope_deg < 10:
        return "upland", "gentle"
    elif elevation_m < 1000 and slope_deg < 20:
        return "hill", "moderate"
    elif elevation_m < 2000 and slope_deg < 30:
        return "mountain_foothill", "moderate"
    elif elevation_m < 3000 and slope_deg < 40:
        return "mountain", "steep"
    elif elevation_m < 5000 and slope_deg < 50:
        return "high_mountain", "steep"
    else:
        return "alpine", "cliff"


def estimate_elevation(lat, lng):
    """Simple elevation estimation."""
    # Use a hash-based approach for deterministic results
    h = int(abs(lat * 1000) + abs(lng * 1000)) % 10000

    # Base elevation from latitude (rough approximation)
    if abs(lat) > 70:
        base = 500 + (h % 2000)
    elif abs(lat) > 50:
        base = 100 + (h % 1000)
    elif abs(lat) > 30:
        base = 50 + (h % 500)
    else:
        base = 10 + (h % 200)

    # Adjust for known mountain ranges (very rough)
    # Himalayas
    if 25 <= lat <= 35 and 70 <= lng <= 100:
        base = max(base, 2000 + (h % 6000))
    # Andes
    elif -40 <= lat <= 10 and -80 <= lng <= -65:
        base = max(base, 1000 + (h % 5000))
    # Rockies
    elif 30 <= lat <= 60 and -120 <= lng <= -100:
        base = max(base, 500 + (h % 3000))
    # Alps
    elif 44 <= lat <= 48 and 6 <= lng <= 16:
        base = max(base, 500 + (h % 3500))
    # Coastal areas (near 0 elevation)
    elif abs(lat) < 5 and (lng < -60 or lng > 100):
        base = min(base, 50)

    return base


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
        elevation_m = estimate_elevation(lat, lng)

        # Simulate slope based on elevation
        h = int(abs(lat * 100) + abs(lng * 100)) % 1000
        if elevation_m < 100:
            slope_deg = round((h % 50) / 10, 1)
        elif elevation_m < 500:
            slope_deg = round(5 + (h % 100) / 10, 1)
        elif elevation_m < 2000:
            slope_deg = round(10 + (h % 200) / 10, 1)
        else:
            slope_deg = round(20 + (h % 300) / 10, 1)

        # Aspect (direction slope faces)
        aspect_deg = (h * 37) % 360

        terrain_type, roughness = classify_terrain(elevation_m, slope_deg)

        # Determine aspect direction
        aspect_dirs = ["N", "NE", "E", "SE", "S", "SW", "W", "NW"]
        aspect_dir = aspect_dirs[round(aspect_deg / 45) % 8]

        return {
            "ok": True,
            "result": {
                "lat": lat,
                "lng": lng,
                "elevation_m": elevation_m,
                "elevation_ft": round(elevation_m * 3.28084, 1),
                "slope_degrees": slope_deg,
                "aspect_degrees": aspect_deg,
                "aspect_direction": aspect_dir,
                "terrain_type": terrain_type,
                "roughness": roughness,
                "note": "Simulated terrain data - use DEM/SRTM for accurate values"
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"terrain data query failed: {str(e)}"}
