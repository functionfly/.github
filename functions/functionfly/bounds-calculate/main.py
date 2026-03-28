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

    if len(coordinates) < 1:
        return {"ok": False, "error": "coordinates must contain at least one point"}

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

        lats = [p[0] for p in parsed]
        lngs = [p[1] for p in parsed]

        min_lat = min(lats)
        max_lat = max(lats)
        min_lng = min(lngs)
        max_lng = max(lngs)

        center_lat = (min_lat + max_lat) / 2
        center_lng = (min_lng + max_lng) / 2

        # Calculate width and height in km
        width_km = haversine_km(center_lat, min_lng, center_lat, max_lng)
        height_km = haversine_km(min_lat, center_lng, max_lat, center_lng)

        return {
            "ok": True,
            "result": {
                "min_lat": round(min_lat, 6),
                "min_lng": round(min_lng, 6),
                "max_lat": round(max_lat, 6),
                "max_lng": round(max_lng, 6),
                "center_lat": round(center_lat, 6),
                "center_lng": round(center_lng, 6),
                "width_km": round(width_km, 4),
                "height_km": round(height_km, 4),
                "point_count": len(parsed)
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"bounds calculation failed: {str(e)}"}
