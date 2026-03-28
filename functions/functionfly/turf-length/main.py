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

    coordinates = event.get("coordinates")
    if coordinates is None:
        return {"ok": False, "error": "coordinates is required"}

    if not isinstance(coordinates, list):
        return {"ok": False, "error": "coordinates must be an array"}

    if len(coordinates) < 2:
        return {"ok": False, "error": "coordinates must contain at least 2 points"}

    unit = event.get("unit", "km")
    valid_units = ["km", "miles", "meters"]
    if unit not in valid_units:
        return {"ok": False, "error": f"unit must be one of: {', '.join(valid_units)}"}

    try:
        parsed = []
        for i, coord in enumerate(coordinates):
            if not isinstance(coord, (list, tuple)) or len(coord) < 2:
                return {"ok": False, "error": f"coordinate at index {i} must be [lat, lng]"}
            lat = float(coord[0])
            lng = float(coord[1])
            if not (-90 <= lat <= 90):
                return {"ok": False, "error": f"lat at index {i} must be between -90 and 90"}
            if not (-180 <= lng <= 180):
                return {"ok": False, "error": f"lng at index {i} must be between -180 and 180"}
            parsed.append((lat, lng))

        total_km = 0.0
        for i in range(len(parsed) - 1):
            total_km += haversine_km(parsed[i][0], parsed[i][1], parsed[i+1][0], parsed[i+1][1])

        total_miles = total_km * 0.621371
        total_meters = total_km * 1000

        unit_map = {"km": total_km, "miles": total_miles, "meters": total_meters}

        return {
            "ok": True,
            "result": {
                "length": round(unit_map[unit], 4),
                "unit": unit,
                "length_km": round(total_km, 4),
                "length_miles": round(total_miles, 4),
                "length_meters": round(total_meters, 2),
                "segment_count": len(parsed) - 1
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"length calculation failed: {str(e)}"}
