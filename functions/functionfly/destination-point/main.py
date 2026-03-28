import math


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}

    for field in ["lat", "lng", "bearing", "distance"]:
        if event.get(field) is None:
            return {"ok": False, "error": f"{field} is required"}

    try:
        lat = float(event["lat"])
        lng = float(event["lng"])
        brng = float(event["bearing"])
        distance = float(event["distance"])
    except (TypeError, ValueError):
        return {"ok": False, "error": "lat, lng, bearing, distance must be numbers"}

    if not (-90 <= lat <= 90):
        return {"ok": False, "error": "lat must be between -90 and 90"}
    if not (-180 <= lng <= 180):
        return {"ok": False, "error": "lng must be between -180 and 180"}
    if distance < 0:
        return {"ok": False, "error": "distance must be non-negative"}

    unit = event.get("unit", "km")
    valid_units = ["km", "miles", "meters"]
    if unit not in valid_units:
        return {"ok": False, "error": f"unit must be one of: {', '.join(valid_units)}"}

    try:
        # Convert distance to km
        if unit == "miles":
            dist_km = distance * 1.60934
        elif unit == "meters":
            dist_km = distance / 1000.0
        else:
            dist_km = distance

        R = 6371.0  # Earth radius in km
        delta = dist_km / R  # angular distance in radians

        theta = math.radians(brng)
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
        dest_lng = (math.degrees(lambda2) + 540) % 360 - 180  # normalize to -180..+180

        return {
            "ok": True,
            "result": {
                "lat": round(dest_lat, 6),
                "lng": round(dest_lng, 6),
                "bearing": brng,
                "distance": distance,
                "unit": unit
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"destination point calculation failed: {str(e)}"}
