def point_in_polygon(lng, lat, polygon):
    """Ray casting algorithm. polygon is list of [lng, lat]."""
    n = len(polygon)
    inside = False
    j = n - 1
    for i in range(n):
        xi, yi = polygon[i][0], polygon[i][1]
        xj, yj = polygon[j][0], polygon[j][1]
        if ((yi > lat) != (yj > lat)) and (lng < (xj - xi) * (lat - yi) / (yj - yi) + xi):
            inside = not inside
        j = i
    return inside


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}

    point = event.get("point")
    polygon = event.get("polygon")

    if point is None:
        return {"ok": False, "error": "point is required"}
    if polygon is None:
        return {"ok": False, "error": "polygon is required"}

    if not isinstance(point, (list, tuple)) or len(point) < 2:
        return {"ok": False, "error": "point must be [lng, lat]"}

    if not isinstance(polygon, list) or len(polygon) < 3:
        return {"ok": False, "error": "polygon must be an array of at least 3 points"}

    try:
        pt_lng = float(point[0])
        pt_lat = float(point[1])
    except (TypeError, ValueError):
        return {"ok": False, "error": "point coordinates must be numbers"}

    if not (-90 <= pt_lat <= 90):
        return {"ok": False, "error": "point lat must be between -90 and 90"}
    if not (-180 <= pt_lng <= 180):
        return {"ok": False, "error": "point lng must be between -180 and 180"}

    try:
        parsed_polygon = []
        for i, coord in enumerate(polygon):
            if not isinstance(coord, (list, tuple)) or len(coord) < 2:
                return {"ok": False, "error": f"polygon coordinate at index {i} must be [lng, lat]"}
            lng = float(coord[0])
            lat = float(coord[1])
            parsed_polygon.append([lng, lat])

        within = point_in_polygon(pt_lng, pt_lat, parsed_polygon)

        return {
            "ok": True,
            "result": {
                "within": within
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"within check failed: {str(e)}"}
