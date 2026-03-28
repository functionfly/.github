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


def segments_intersect(p1, p2, p3, p4):
    """Check if segment p1-p2 intersects with segment p3-p4."""
    d1x = p2[0] - p1[0]
    d1y = p2[1] - p1[1]
    d2x = p4[0] - p3[0]
    d2y = p4[1] - p3[1]
    denom = d1x * d2y - d1y * d2x
    if abs(denom) < 1e-10:
        return False
    dx = p3[0] - p1[0]
    dy = p3[1] - p1[1]
    t = (dx * d2y - dy * d2x) / denom
    u = (dx * d1y - dy * d1x) / denom
    return 0 <= t <= 1 and 0 <= u <= 1


def parse_polygon(poly, name):
    if not isinstance(poly, list) or len(poly) < 3:
        return None, f"{name} must be an array of at least 3 points"
    pts = []
    for i, coord in enumerate(poly):
        if not isinstance(coord, (list, tuple)) or len(coord) < 2:
            return None, f"{name} coordinate at index {i} must be [lng, lat]"
        try:
            lng = float(coord[0])
            lat = float(coord[1])
        except (TypeError, ValueError):
            return None, f"{name} coordinate at index {i} must be numbers"
        pts.append([lng, lat])
    return pts, None


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}

    poly1_raw = event.get("polygon1")
    poly2_raw = event.get("polygon2")

    if poly1_raw is None:
        return {"ok": False, "error": "polygon1 is required"}
    if poly2_raw is None:
        return {"ok": False, "error": "polygon2 is required"}

    poly1, err = parse_polygon(poly1_raw, "polygon1")
    if err:
        return {"ok": False, "error": err}

    poly2, err = parse_polygon(poly2_raw, "polygon2")
    if err:
        return {"ok": False, "error": err}

    try:
        # Check if any vertex of poly1 is inside poly2
        for pt in poly1:
            if point_in_polygon(pt[0], pt[1], poly2):
                return {"ok": True, "result": {"intersects": True}}

        # Check if any vertex of poly2 is inside poly1
        for pt in poly2:
            if point_in_polygon(pt[0], pt[1], poly1):
                return {"ok": True, "result": {"intersects": True}}

        # Check if any edges intersect
        n1 = len(poly1)
        n2 = len(poly2)
        for i in range(n1):
            j = (i + 1) % n1
            p1 = poly1[i]
            p2 = poly1[j]
            for k in range(n2):
                l = (k + 1) % n2
                p3 = poly2[k]
                p4 = poly2[l]
                if segments_intersect(p1, p2, p3, p4):
                    return {"ok": True, "result": {"intersects": True}}

        return {"ok": True, "result": {"intersects": False}}

    except Exception as e:
        return {"ok": False, "error": f"intersects check failed: {str(e)}"}
