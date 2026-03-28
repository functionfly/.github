def segments_intersect_point(p1, p2, p3, p4):
    """Check if segment p1-p2 intersects with segment p3-p4.
    Returns intersection point or None.
    """
    d1x = p2[0] - p1[0]
    d1y = p2[1] - p1[1]
    d2x = p4[0] - p3[0]
    d2y = p4[1] - p3[1]
    denom = d1x * d2y - d1y * d2x

    if abs(denom) < 1e-10:
        return None

    dx = p3[0] - p1[0]
    dy = p3[1] - p1[1]
    t = (dx * d2y - dy * d2x) / denom
    u = (dx * d1y - dy * d1x) / denom

    # Use strict interior intersection (not at endpoints)
    if 1e-10 < t < 1 - 1e-10 and 1e-10 < u < 1 - 1e-10:
        ix = p1[0] + t * d1x
        iy = p1[1] + t * d1y
        return [round(ix, 6), round(iy, 6)]

    return None


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}

    coordinates = event.get("coordinates")
    if coordinates is None:
        return {"ok": False, "error": "coordinates is required"}

    if not isinstance(coordinates, list) or len(coordinates) < 3:
        return {"ok": False, "error": "coordinates must be an array of at least 3 points"}

    try:
        parsed = []
        for i, coord in enumerate(coordinates):
            if not isinstance(coord, (list, tuple)) or len(coord) < 2:
                return {"ok": False, "error": f"coordinate at index {i} must be [lng, lat]"}
            lng = float(coord[0])
            lat = float(coord[1])
            parsed.append((lng, lat))

        n = len(parsed)
        kinks = []
        seen = set()

        # Check all non-adjacent segment pairs
        for i in range(n - 1):
            for j in range(i + 2, n - 1):
                # Skip adjacent segments
                if i == 0 and j == n - 2:
                    continue
                pt = segments_intersect_point(
                    parsed[i], parsed[i+1],
                    parsed[j], parsed[j+1]
                )
                if pt:
                    key = (round(pt[0], 4), round(pt[1], 4))
                    if key not in seen:
                        seen.add(key)
                        kinks.append(pt)

        return {
            "ok": True,
            "result": {
                "kinks": kinks,
                "kink_count": len(kinks),
                "is_simple": len(kinks) == 0
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"kinks detection failed: {str(e)}"}
