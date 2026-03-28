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
    # Super triangle
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


def clip_polygon_to_bbox(polygon, bbox):
    """Clip polygon to bounding box using Sutherland-Hodgman."""
    min_x, min_y, max_x, max_y = bbox

    def inside(p, edge):
        if edge == 'left': return p[0] >= min_x
        if edge == 'right': return p[0] <= max_x
        if edge == 'bottom': return p[1] >= min_y
        if edge == 'top': return p[1] <= max_y

    def intersect(p1, p2, edge):
        x1, y1 = p1
        x2, y2 = p2
        if edge == 'left':
            t = (min_x - x1) / (x2 - x1) if x2 != x1 else 0
            return (min_x, y1 + t * (y2 - y1))
        if edge == 'right':
            t = (max_x - x1) / (x2 - x1) if x2 != x1 else 0
            return (max_x, y1 + t * (y2 - y1))
        if edge == 'bottom':
            t = (min_y - y1) / (y2 - y1) if y2 != y1 else 0
            return (x1 + t * (x2 - x1), min_y)
        if edge == 'top':
            t = (max_y - y1) / (y2 - y1) if y2 != y1 else 0
            return (x1 + t * (x2 - x1), max_y)

    output = polygon
    for edge in ['left', 'right', 'bottom', 'top']:
        if not output:
            return []
        input_list = output
        output = []
        for i in range(len(input_list)):
            current = input_list[i]
            previous = input_list[i - 1]
            if inside(current, edge):
                if not inside(previous, edge):
                    output.append(intersect(previous, current, edge))
                output.append(current)
            elif inside(previous, edge):
                output.append(intersect(previous, current, edge))
    return output


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}

    points_raw = event.get("points")
    if points_raw is None:
        return {"ok": False, "error": "points is required"}

    if not isinstance(points_raw, list) or len(points_raw) < 3:
        return {"ok": False, "error": "points must be an array of at least 3 points"}

    if len(points_raw) > 100:
        return {"ok": False, "error": "points array is limited to 100 points"}

    try:
        parsed = []
        for i, pt in enumerate(points_raw):
            if not isinstance(pt, (list, tuple)) or len(pt) < 2:
                return {"ok": False, "error": f"point at index {i} must be [lng, lat]"}
            lng = float(pt[0])
            lat = float(pt[1])
            parsed.append((lng, lat))

        # Default bbox
        bbox_raw = event.get("bbox")
        if bbox_raw:
            if not isinstance(bbox_raw, (list, tuple)) or len(bbox_raw) < 4:
                return {"ok": False, "error": "bbox must be [minLng, minLat, maxLng, maxLat]"}
            bbox = [float(bbox_raw[0]), float(bbox_raw[1]), float(bbox_raw[2]), float(bbox_raw[3])]
        else:
            lngs = [p[0] for p in parsed]
            lats = [p[1] for p in parsed]
            margin = max((max(lngs) - min(lngs)) * 0.2, 1.0)
            bbox = [min(lngs) - margin, min(lats) - margin, max(lngs) + margin, max(lats) + margin]

        # Compute Delaunay triangulation
        triangles = bowyer_watson(parsed)

        # Build Voronoi cells: for each input point, find all triangles containing it
        # and collect their circumcenters to form the Voronoi cell
        cells = []
        for pt in parsed:
            cell_centers = []
            for tri in triangles:
                if pt in tri:
                    cc = circumcenter(tri[0][0], tri[0][1], tri[1][0], tri[1][1], tri[2][0], tri[2][1])
                    if cc:
                        cell_centers.append(cc)

            if len(cell_centers) < 2:
                # Fallback: create a small square around the point
                d = 0.5
                cell_centers = [
                    (pt[0] - d, pt[1] - d),
                    (pt[0] + d, pt[1] - d),
                    (pt[0] + d, pt[1] + d),
                    (pt[0] - d, pt[1] + d)
                ]

            # Sort cell centers by angle around the input point
            import math
            cell_centers.sort(key=lambda c: math.atan2(c[1] - pt[1], c[0] - pt[0]))

            # Clip to bbox
            clipped = clip_polygon_to_bbox(cell_centers, bbox)
            if clipped:
                # Close the polygon
                poly = [[round(c[0], 6), round(c[1], 6)] for c in clipped]
                if poly[0] != poly[-1]:
                    poly.append(poly[0])
                cells.append(poly)

        return {
            "ok": True,
            "result": {
                "cells": cells,
                "point_count": len(parsed),
                "triangle_count": len(triangles)
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"Voronoi generation failed: {str(e)}"}
