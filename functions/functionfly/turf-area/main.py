import math


def ring_area(coords):
    """Calculate area of a ring using spherical excess formula.
    coords: list of [lng, lat] in degrees (GeoJSON order)
    Returns area in square meters.
    """
    R = 6371008.8  # Earth radius in meters (WGS84 mean)
    n = len(coords)
    if n < 3:
        return 0.0

    total = 0.0
    for i in range(n):
        j = (i + 1) % n
        lng1 = math.radians(coords[i][0])
        lat1 = math.radians(coords[i][1])
        lng2 = math.radians(coords[j][0])
        lat2 = math.radians(coords[j][1])
        total += (lng2 - lng1) * (2 + math.sin(lat1) + math.sin(lat2))

    return abs(total * R * R / 2.0)


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}

    coordinates = event.get("coordinates")
    if coordinates is None:
        return {"ok": False, "error": "coordinates is required"}

    if not isinstance(coordinates, list):
        return {"ok": False, "error": "coordinates must be an array"}

    if len(coordinates) < 3:
        return {"ok": False, "error": "coordinates must contain at least 3 points"}

    try:
        parsed = []
        for i, coord in enumerate(coordinates):
            if not isinstance(coord, (list, tuple)) or len(coord) < 2:
                return {"ok": False, "error": f"coordinate at index {i} must be [lng, lat]"}
            lng = float(coord[0])
            lat = float(coord[1])
            if not (-90 <= lat <= 90):
                return {"ok": False, "error": f"lat at index {i} must be between -90 and 90"}
            if not (-180 <= lng <= 180):
                return {"ok": False, "error": f"lng at index {i} must be between -180 and 180"}
            parsed.append([lng, lat])

        area_sq_meters = ring_area(parsed)
        area_sq_km = area_sq_meters / 1_000_000
        area_hectares = area_sq_km * 100
        area_sq_miles = area_sq_km * 0.386102

        return {
            "ok": True,
            "result": {
                "area_sq_meters": round(area_sq_meters, 2),
                "area_sq_km": round(area_sq_km, 6),
                "area_hectares": round(area_hectares, 4),
                "area_sq_miles": round(area_sq_miles, 6)
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"area calculation failed: {str(e)}"}
