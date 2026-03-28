import math


def haversine_km(lat1, lng1, lat2, lng2):
    R = 6371.0
    phi1 = math.radians(lat1)
    phi2 = math.radians(lat2)
    dphi = math.radians(lat2 - lat1)
    dlambda = math.radians(lng2 - lng1)
    a = math.sin(dphi / 2) ** 2 + math.cos(phi1) * math.cos(phi2) * math.sin(dlambda / 2) ** 2
    return R * 2 * math.atan2(math.sqrt(a), math.sqrt(1 - a))


def spherical_polygon_area(coords):
    """Calculate area of a spherical polygon using the spherical excess formula.
    coords: list of (lat, lng) in degrees
    Returns area in square meters.
    """
    R = 6371000.0  # Earth radius in meters
    n = len(coords)
    if n < 3:
        return 0.0

    # Ensure polygon is closed
    if coords[0] != coords[-1]:
        coords = coords + [coords[0]]

    # Use the shoelace formula adapted for spherical coordinates
    # Convert to radians
    pts = [(math.radians(lat), math.radians(lng)) for lat, lng in coords]

    total = 0.0
    for i in range(len(pts) - 1):
        phi1, lambda1 = pts[i]
        phi2, lambda2 = pts[i + 1]
        total += (lambda2 - lambda1) * (2 + math.sin(phi1) + math.sin(phi2))

    area = abs(total * R * R / 2.0)
    return area


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}

    coordinates = event.get("coordinates")
    if coordinates is None:
        return {"ok": False, "error": "coordinates is required"}

    if not isinstance(coordinates, list):
        return {"ok": False, "error": "coordinates must be an array"}

    if len(coordinates) < 3:
        return {"ok": False, "error": "coordinates must contain at least 3 points to form a polygon"}

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

        # Calculate area
        area_sq_meters = spherical_polygon_area(parsed)
        area_sq_km = area_sq_meters / 1_000_000
        area_sq_miles = area_sq_km * 0.386102
        area_hectares = area_sq_km * 100

        # Calculate perimeter
        perimeter_km = 0.0
        n = len(parsed)
        for i in range(n):
            j = (i + 1) % n
            perimeter_km += haversine_km(parsed[i][0], parsed[i][1], parsed[j][0], parsed[j][1])

        return {
            "ok": True,
            "result": {
                "area_sq_km": round(area_sq_km, 6),
                "area_sq_miles": round(area_sq_miles, 6),
                "area_sq_meters": round(area_sq_meters, 2),
                "area_hectares": round(area_hectares, 4),
                "perimeter_km": round(perimeter_km, 4)
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"polygon area calculation failed: {str(e)}"}
