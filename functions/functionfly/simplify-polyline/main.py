import math


def perpendicular_distance(point, line_start, line_end):
    """Calculate perpendicular distance from point to line segment."""
    x0, y0 = point[1], point[0]  # lng, lat
    x1, y1 = line_start[1], line_start[0]
    x2, y2 = line_end[1], line_end[0]

    dx = x2 - x1
    dy = y2 - y1

    if dx == 0 and dy == 0:
        # Line start and end are the same point
        return math.sqrt((x0 - x1) ** 2 + (y0 - y1) ** 2)

    # Distance from point to line
    t = ((x0 - x1) * dx + (y0 - y1) * dy) / (dx * dx + dy * dy)
    t = max(0, min(1, t))

    nearest_x = x1 + t * dx
    nearest_y = y1 + t * dy

    return math.sqrt((x0 - nearest_x) ** 2 + (y0 - nearest_y) ** 2)


def rdp(points, tolerance):
    """Ramer-Douglas-Peucker algorithm."""
    if len(points) <= 2:
        return points

    # Find the point with the maximum distance
    max_dist = 0
    max_idx = 0

    for i in range(1, len(points) - 1):
        dist = perpendicular_distance(points[i], points[0], points[-1])
        if dist > max_dist:
            max_dist = dist
            max_idx = i

    if max_dist > tolerance:
        # Recursively simplify
        left = rdp(points[:max_idx + 1], tolerance)
        right = rdp(points[max_idx:], tolerance)
        return left[:-1] + right
    else:
        return [points[0], points[-1]]


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

    tolerance = event.get("tolerance", 0.0001)
    try:
        tolerance = float(tolerance)
    except (TypeError, ValueError):
        return {"ok": False, "error": "tolerance must be a number"}

    if tolerance <= 0:
        return {"ok": False, "error": "tolerance must be positive"}

    try:
        parsed = []
        for i, coord in enumerate(coordinates):
            if not isinstance(coord, (list, tuple)) or len(coord) < 2:
                return {"ok": False, "error": f"coordinate at index {i} must be [lat, lng]"}
            lat = float(coord[0])
            lng = float(coord[1])
            parsed.append((lat, lng))

        original_count = len(parsed)
        simplified = rdp(parsed, tolerance)
        simplified_count = len(simplified)

        reduction = 0.0
        if original_count > 0:
            reduction = round((1 - simplified_count / original_count) * 100, 2)

        return {
            "ok": True,
            "result": {
                "coordinates": [[p[0], p[1]] for p in simplified],
                "original_count": original_count,
                "simplified_count": simplified_count,
                "reduction_percent": reduction,
                "tolerance": tolerance
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"polyline simplification failed: {str(e)}"}
