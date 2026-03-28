def point_in_polygon_ray_cast(lat, lng, polygon):
    """Ray casting algorithm to check if point is inside polygon.
    Returns (inside, on_boundary).
    """
    n = len(polygon)
    inside = False
    on_boundary = False

    j = n - 1
    for i in range(n):
        xi, yi = polygon[i][1], polygon[i][0]  # lng, lat
        xj, yj = polygon[j][1], polygon[j][0]

        # Check if point is on the edge
        # Using cross product and dot product
        if abs((yi - yj) * lng - (xi - xj) * lat + xi * yj - yi * xj) < 1e-10:
            # Point is on the line containing this edge
            min_x = min(xi, xj)
            max_x = max(xi, xj)
            min_y = min(yi, yj)
            max_y = max(yi, yj)
            if min_x <= lng <= max_x and min_y <= lat <= max_y:
                on_boundary = True
                return inside, on_boundary

        # Ray casting
        if ((yi > lat) != (yj > lat)) and (lng < (xj - xi) * (lat - yi) / (yj - yi) + xi):
            inside = not inside

        j = i

    return inside, on_boundary


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
        return {"ok": False, "error": "point must be [lat, lng]"}

    if not isinstance(polygon, list):
        return {"ok": False, "error": "polygon must be an array"}

    if len(polygon) < 3:
        return {"ok": False, "error": "polygon must contain at least 3 points"}

    try:
        pt_lat = float(point[0])
        pt_lng = float(point[1])
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
                return {"ok": False, "error": f"polygon coordinate at index {i} must be [lat, lng]"}
            lat = float(coord[0])
            lng = float(coord[1])
            if not (-90 <= lat <= 90):
                return {"ok": False, "error": f"polygon lat at index {i} must be between -90 and 90"}
            if not (-180 <= lng <= 180):
                return {"ok": False, "error": f"polygon lng at index {i} must be between -180 and 180"}
            parsed_polygon.append((lat, lng))

        inside, on_boundary = point_in_polygon_ray_cast(pt_lat, pt_lng, parsed_polygon)

        return {
            "ok": True,
            "result": {
                "inside": inside or on_boundary,
                "strictly_inside": inside and not on_boundary,
                "on_boundary": on_boundary
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"point-in-polygon check failed: {str(e)}"}
