import math


def destination_point(lat, lng, bearing_deg, dist_km):
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
    return round(dest_lng, 6), round(dest_lat, 6)  # GeoJSON order: [lng, lat]


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}

    geojson = event.get("geojson")
    radius = event.get("radius")

    if geojson is None:
        return {"ok": False, "error": "geojson is required"}
    if radius is None:
        return {"ok": False, "error": "radius is required"}

    if not isinstance(geojson, dict):
        return {"ok": False, "error": "geojson must be an object"}

    try:
        radius = float(radius)
    except (TypeError, ValueError):
        return {"ok": False, "error": "radius must be a number"}

    if radius <= 0:
        return {"ok": False, "error": "radius must be positive"}

    unit = event.get("unit", "km")
    valid_units = ["km", "miles", "meters"]
    if unit not in valid_units:
        return {"ok": False, "error": f"unit must be one of: {', '.join(valid_units)}"}

    steps = event.get("steps", 64)
    try:
        steps = max(8, min(360, int(steps)))
    except (TypeError, ValueError):
        steps = 64

    try:
        # Extract point coordinates
        gtype = geojson.get("type")
        if gtype == "Feature":
            geometry = geojson.get("geometry", {})
            gtype = geometry.get("type")
            coords = geometry.get("coordinates")
        else:
            coords = geojson.get("coordinates")

        if gtype != "Point":
            return {"ok": False, "error": "geojson must be a Point geometry or Feature with Point geometry"}

        if not isinstance(coords, (list, tuple)) or len(coords) < 2:
            return {"ok": False, "error": "Point coordinates must be [lng, lat]"}

        lng = float(coords[0])
        lat = float(coords[1])

        # Convert radius to km
        if unit == "miles":
            radius_km = radius * 1.60934
        elif unit == "meters":
            radius_km = radius / 1000.0
        else:
            radius_km = radius

        # Generate circle polygon
        ring = []
        for i in range(steps):
            bearing = (360.0 / steps) * i
            pt_lng, pt_lat = destination_point(lat, lng, bearing, radius_km)
            ring.append([pt_lng, pt_lat])

        # Close the ring
        ring.append(ring[0])

        return {
            "ok": True,
            "result": {
                "type": "Polygon",
                "coordinates": [ring]
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"GeoJSON buffer failed: {str(e)}"}
