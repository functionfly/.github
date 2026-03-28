def circumcenter(ax, ay, bx, by, cx, cy):
    """Calculate circumcenter of triangle."""
    D = 2 * (ax * (by - cy) + bx * (cy - ay) + cx * (ay - by))
    if abs(D) < 1e-10:
        return None
    ux = ((ax**2 + ay**2) * (by - cy) + (bx**2 + by**2) * (cy - ay) + (cx**2 + cy**2) * (ay - by)) / D
    uy = ((ax**2 + ay**2) * (cx - bx) + (bx**2 + by**2) * (ax - cx) + (cx**2 + cy**2) * (bx - ax)) / D
    return ux, uy


def bowyer_watson(points):
    """Bowyer-Watson Delaunay triangulation."""
    min_x = min(p[0] for p in points) - 10
    min_y = min(p[1] for p in points) - 10
    max_x = max(p[0] for p in points) + 10
    max_y = max(p[1] for p in points) + 10
    dx = max_x - min_x
    dy = max_y - min_y
    delta_max = max(dx, dy)
    mid_x = (min_x + max_x) / 2
    mid_y = (min_y + max_y) / 2

    p1 = (mid_x - 20 * delta_max, mid_y - delta_max)
    p2 = (mid_x, mid_y + 20 * delta_max)
    p3 = (mid_x + 20 * delta_max, mid_y - delta_max)

    triangles = [(p1, p2, p3)]

    for point in points:
        bad_triangles = []
        for tri in triangles:
            ax, ay = tri[0]
            bx, by = tri[1]
            cx, cy = tri[2]
            cc = circumcenter(ax, ay, bx, by, cx, cy)
            if cc is None:
                continue
            ux, uy = cc
            r2 = (ax - ux)**2 + (ay - uy)**2
            d2 = (point[0] - ux)**2 + (point[1] - uy)**2
            if d2 <= r2 + 1e-10:
                bad_triangles.append(tri)

        boundary = []
        for tri in bad_triangles:
            edges = [(tri[0], tri[1]), (tri[1], tri[2]), (tri[2], tri[0])]
            for edge in edges:
                shared = False
                for other in bad_triangles:
                    if other == tri:
                        continue
                    other_edges = [(other[0], other[1]), (other[1], other[2]), (other[2], other[0])]
                    for oe in other_edges:
                        if (edge[0] == oe[1] and edge[1] == oe[0]) or (edge[0] == oe[0] and edge[1] == oe[1]):
                            shared = True
                            break
                    if shared:
                        break
                if not shared:
                    boundary.append(edge)

        for tri in bad_triangles:
            triangles.remove(tri)

        for edge in boundary:
            triangles.append((edge[0], edge[1], point))

    # Remove triangles that share vertices with super triangle
    super_verts = {p1, p2, p3}
    triangles = [t for t in triangles if not (t[0] in super_verts or t[1] in super_verts or t[2] in super_verts)]
    return triangles


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}

    points_raw = event.get("points")
    if points_raw is None:
        return {"ok": False, "error": "points is required"}

    if not isinstance(points_raw, list) or len(points_raw) < 3:
        return {"ok": False, "error": "points must be an array of at least 3 points"}

    if len(points_raw) > 200:
        return {"ok": False, "error": "points array is limited to 200 points"}

    try:
        parsed = []
        for i, pt in enumerate(points_raw):
            if not isinstance(pt, (list, tuple)) or len(pt) < 2:
                return {"ok": False, "error": f"point at index {i} must be [lng, lat]"}
            lng = float(pt[0])
            lat = float(pt[1])
            parsed.append((lng, lat))

        triangles = bowyer_watson(parsed)

        result_triangles = []
        for tri in triangles:
            result_triangles.append([
                [round(tri[0][0], 6), round(tri[0][1], 6)],
                [round(tri[1][0], 6), round(tri[1][1], 6)],
                [round(tri[2][0], 6), round(tri[2][1], 6)]
            ])

        return {
            "ok": True,
            "result": {
                "triangles": result_triangles,
                "triangle_count": len(result_triangles),
                "point_count": len(parsed)
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"TIN generation failed: {str(e)}"}
