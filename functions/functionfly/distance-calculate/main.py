import math


def haversine_km(lat1, lng1, lat2, lng2):
    R = 6371.0
    phi1 = math.radians(lat1)
    phi2 = math.radians(lat2)
    dphi = math.radians(lat2 - lat1)
    dlambda = math.radians(lng2 - lng1)
    a = math.sin(dphi / 2) ** 2 + math.cos(phi1) * math.cos(phi2) * math.sin(dlambda / 2) ** 2
    return R * 2 * math.atan2(math.sqrt(a), math.sqrt(1 - a))


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}

    for field in ["lat1", "lng1", "lat2", "lng2"]:
        if event.get(field) is None:
            return {"ok": False, "error": f"{field} is required"}

    try:
        lat1 = float(event["lat1"])
        lng1 = float(event["lng1"])
        lat2 = float(event["lat2"])
        lng2 = float(event["lng2"])
    except (TypeError, ValueError):
        return {"ok": False, "error": "lat1, lng1, lat2, lng2 must be numbers"}

    if not (-90 <= lat1 <= 90):
        return {"ok": False, "error": "lat1 must be between -90 and 90"}
    if not (-180 <= lng1 <= 180):
        return {"ok": False, "error": "lng1 must be between -180 and 180"}
    if not (-90 <= lat2 <= 90):
        return {"ok": False, "error": "lat2 must be between -90 and 90"}
    if not (-180 <= lng2 <= 180):
        return {"ok": False, "error": "lng2 must be between -180 and 180"}

    unit = event.get("unit", "km")
    valid_units = ["km", "miles", "meters", "feet"]
    if unit not in valid_units:
        return {"ok": False, "error": f"unit must be one of: {', '.join(valid_units)}"}

    try:
        dist_km = haversine_km(lat1, lng1, lat2, lng2)
        dist_miles = dist_km * 0.621371
        dist_meters = dist_km * 1000
        dist_feet = dist_meters * 3.28084

        unit_map = {
            "km": dist_km,
            "miles": dist_miles,
            "meters": dist_meters,
            "feet": dist_feet
        }

        return {
            "ok": True,
            "result": {
                "distance": round(unit_map[unit], 4),
                "unit": unit,
                "distance_km": round(dist_km, 4),
                "distance_miles": round(dist_miles, 4),
                "distance_meters": round(dist_meters, 2),
                "distance_feet": round(dist_feet, 2)
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"distance calculation failed: {str(e)}"}
