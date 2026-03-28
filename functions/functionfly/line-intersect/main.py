def cross_product_2d(o, a, b):
    """2D cross product of OA and OB vectors."""
    return (a[0] - o[0]) * (b[1] - o[1]) - (a[1] - o[1]) * (b[0] - o[0])


def segments_intersect(p1, p2, p3, p4):
    """Check if segment p1-p2 intersects with segment p3-p4.
    Returns (intersects, intersection_point or None).
    Points are (x, y) tuples.
    """
    d1x = p2[0] - p1[0]
    d1y = p2[1] - p1[1]
    d2x = p4[0] - p3[0]
    d2y = p4[1] - p3[1]

    denom = d1x * d2y - d1y * d2x

    if abs(denom) < 1e-10:
        # Lines are parallel or collinear
        return False, None

    dx = p3[0] - p1[0]
    dy = p3[1] - p1[1]

    t = (dx * d2y - dy * d2x) / denom
    u = (dx * d1y - dy * d1x) / denom

    if 0 <= t <= 1 and 0 <= u <= 1:
        ix = p1[0] + t * d1x
        iy = p1[1] + t * d1y
        return True, (ix, iy)

    return False, None


def parse_line(line, name):
    if not isinstance(line, list) or len(line) < 2:
        return None, f"{name} must be an array of at least 2 points"
    pts = []
    for i, pt in enumerate(line):
        if not isinstance(pt, (list, tuple)) or len(pt) < 2:
            return None, f"{name} point at index {i} must be [lat, lng]"
        try:
            lat = float(pt[0])
            lng = float(pt[1])
        except (TypeError, ValueError):
            return None, f"{name} point at index {i} coordinates must be numbers"
        pts.append((lat, lng))
    return pts, None


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}

    line1_raw = event.get("line1")
    line2_raw = event.get("line2")

    if line1_raw is None:
        return {"ok": False, "error": "line1 is required"}
    if line2_raw is None:
        return {"ok": False, "error": "line2 is required"}

    line1, err = parse_line(line1_raw, "line1")
    if err:
        return {"ok": False, "error": err}

    line2, err = parse_line(line2_raw, "line2")
    if err:
        return {"ok": False, "error": err}

    try:
        # Check all segment pairs for intersection
        # For multi-segment lines, check each pair
        intersections = []
        for i in range(len(line1) - 1):
            for j in range(len(line2) - 1):
                p1 = (line1[i][1], line1[i][0])    # (lng, lat)
                p2 = (line1[i+1][1], line1[i+1][0])
                p3 = (line2[j][1], line2[j][0])
                p4 = (line2[j+1][1], line2[j+1][0])

                intersects, pt = segments_intersect(p1, p2, p3, p4)
                if intersects and pt:
                    intersections.append({"lat": round(pt[1], 6), "lng": round(pt[0], 6)})

        if intersections:
            return {
                "ok": True,
                "result": {
                    "intersects": True,
                    "intersection": intersections[0],
                    "all_intersections": intersections
                }
            }
        else:
            return {
                "ok": True,
                "result": {
                    "intersects": False,
                    "intersection": None,
                    "all_intersections": []
                }
            }

    except Exception as e:
        return {"ok": False, "error": f"line intersection check failed: {str(e)}"}
