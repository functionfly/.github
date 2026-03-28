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

    target = event.get("target")
    points = event.get("points")

    if target is None:
        return {"ok": False, "error": "target is required"}
    if points is None:
        return {"ok": False, "error": "points is required"}

    if not isinstance(target, (list, tuple)) or len(target) < 2:
        return {"ok": False, "error": "target must be [lng, lat]"}

    if not isinstance(points, list) or len(points) < 1:
        return {"ok": False, "error": "points must be a non-empty array"}

    try:
        t_lng = float(target[0])
        t_lat = float(target[1])
    except (TypeError, ValueError):
        return {"ok": False, "error": "target coordinates must be numbers"}

    if not (-90 <= t_lat <= 90):
        return {"ok": False, "error": "target lat must be between -90 and 90"}
    if not (-180 <= t_lng <= 180):
        return {"ok": False, "error": "target lng must be between -180 and 180"}

    try:
        nearest_idx = None
        nearest_dist = float("inf")
        nearest_pt = None

        for i, pt in enumerate(points):
            if not isinstance(pt, (list, tuple)) or len(pt) < 2:
                return {"ok": False, "error": f"point at index {i} must be [lng, lat]"}
            lng = float(pt[0])
            lat = float(pt[1])
            dist = haversine_km(t_lat, t_lng, lat, lng)
            if dist < nearest_dist:
                nearest_dist = dist
                nearest_idx = i
                nearest_pt = [lng, lat]

        return {
            "ok": True,
            "result": {
                "nearest": nearest_pt,
                "index": nearest_idx,
                "distance_km": round(nearest_dist, 4)
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"nearest point search failed: {str(e)}"}
