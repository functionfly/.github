import math


def destination_point(lat, lng, bearing_deg, dist_km):
    """Calculate destination point given start, bearing, and distance."""
    R = 6371.0
    delta = dist_km / R
    theta = math.radians(bearing_deg)
    phi1 = math.radians(lat)
    lambda1 = math.radians(lng)

    phi2 = math.asin(
        math.sin(phi1) * math.cos(delta) +
        math.cos(phi1) * math.sin(delta) * math.cos(theta)
    )
    lambda2 = lambda1 + math.atan2(
        math.sin(theta) * math.sin(delta) * math.cos(phi1),
        math.cos(delta) - math.sin(phi1) * math.sin(phi2)
    )

    dest_lat = math.degrees(phi2)
    dest_lng = (math.degrees(lambda2) + 540) % 360 - 180
    return round(dest_lat, 6), round(dest_lng, 6)


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}

    lat = event.get("lat")
    lng = event.get("lng")
    radius = event.get("radius")

    if lat is None:
        return {"ok": False, "error": "lat is required"}
    if lng is None:
        return {"ok": False, "error": "lng is required"}
    if radius is None:
        return {"ok": False, "error": "radius is required"}

    try:
        lat = float(lat)
        lng = float(lng)
        radius = float(radius)
    except (TypeError, ValueError):
        return {"ok": False, "error": "lat, lng, radius must be numbers"}

    if not (-90 <= lat <= 90):
        return {"ok": False, "error": "lat must be between -90 and 90"}
    if not (-180 <= lng <= 180):
        return {"ok": False, "error": "lng must be between -180 and 180"}
    if radius <= 0:
        return {"ok": False, "error": "radius must be positive"}

    unit = event.get("unit", "km")
    valid_units = ["km", "miles", "meters"]
    if unit not in valid_units:
        return {"ok": False, "error": f"unit must be one of: {', '.join(valid_units)}"}

    steps = event.get("steps", 64)
    try:
        steps = int(steps)
    except (TypeError, ValueError):
        return {"ok": False, "error": "steps must be an integer"}

    if steps < 8:
        steps = 8
    if steps > 360:
        steps = 360

    try:
        # Convert radius to km
        if unit == "miles":
            radius_km = radius * 1.60934
        elif unit == "meters":
            radius_km = radius / 1000.0
        else:
            radius_km = radius

        # Generate circle polygon
        coordinates = []
        for i in range(steps):
            bearing = (360.0 / steps) * i
            pt_lat, pt_lng = destination_point(lat, lng, bearing, radius_km)
            coordinates.append([pt_lat, pt_lng])

        # Close the polygon
        coordinates.append(coordinates[0])

        return {
            "ok": True,
            "result": {
                "center": {"lat": lat, "lng": lng},
                "radius": radius,
                "unit": unit,
                "radius_km": radius_km,
                "steps": steps,
                "coordinates": coordinates
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"buffer point calculation failed: {str(e)}"}
