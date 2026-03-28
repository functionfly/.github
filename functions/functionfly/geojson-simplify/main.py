import math


def perp_dist(point, line_start, line_end):
    x0, y0 = point[0], point[1]
    x1, y1 = line_start[0], line_start[1]
    x2, y2 = line_end[0], line_end[1]
    dx = x2 - x1
    dy = y2 - y1
    if dx == 0 and dy == 0:
        return math.sqrt((x0 - x1) ** 2 + (y0 - y1) ** 2)
    t = ((x0 - x1) * dx + (y0 - y1) * dy) / (dx * dx + dy * dy)
    t = max(0, min(1, t))
    return math.sqrt((x0 - (x1 + t * dx)) ** 2 + (y0 - (y1 + t * dy)) ** 2)


def rdp(points, tolerance):
    if len(points) <= 2:
        return points
    max_dist = 0
    max_idx = 0
    for i in range(1, len(points) - 1):
        dist = perp_dist(points[i], points[0], points[-1])
        if dist > max_dist:
            max_dist = dist
            max_idx = i
    if max_dist > tolerance:
        left = rdp(points[:max_idx + 1], tolerance)
        right = rdp(points[max_idx:], tolerance)
        return left[:-1] + right
    return [points[0], points[-1]]


def simplify_coords(coords, tolerance):
    if not coords or len(coords) < 2:
        return coords
    simplified = rdp(coords, tolerance)
    return simplified


def simplify_geometry(geom, tolerance):
    gtype = geom.get("type")
    coords = geom.get("coordinates")

    if gtype == "Point":
        return geom

    elif gtype == "MultiPoint":
        return geom

    elif gtype == "LineString":
        simplified = simplify_coords(coords, tolerance)
        return {"type": "LineString", "coordinates": simplified}

    elif gtype == "MultiLineString":
        simplified = [simplify_coords(line, tolerance) for line in coords]
        return {"type": "MultiLineString", "coordinates": simplified}

    elif gtype == "Polygon":
        simplified_rings = []
        for ring in coords:
            s = simplify_coords(ring, tolerance)
            # Ensure ring is closed and has at least 4 points
            if len(s) < 4:
                s = ring  # Keep original if too simplified
            elif s[0] != s[-1]:
                s = s + [s[0]]
            simplified_rings.append(s)
        return {"type": "Polygon", "coordinates": simplified_rings}

    elif gtype == "MultiPolygon":
        simplified_polys = []
        for poly in coords:
            simplified_rings = []
            for ring in poly:
                s = simplify_coords(ring, tolerance)
                if len(s) < 4:
                    s = ring
                elif s[0] != s[-1]:
                    s = s + [s[0]]
                simplified_rings.append(s)
            simplified_polys.append(simplified_rings)
        return {"type": "MultiPolygon", "coordinates": simplified_polys}

    elif gtype == "GeometryCollection":
        geometries = geom.get("geometries", [])
        simplified = [simplify_geometry(g, tolerance) for g in geometries]
        return {"type": "GeometryCollection", "geometries": simplified}

    return geom


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}

    geojson = event.get("geojson")
    if geojson is None:
        return {"ok": False, "error": "geojson is required"}

    if not isinstance(geojson, dict):
        return {"ok": False, "error": "geojson must be an object"}

    tolerance = event.get("tolerance", 0.0001)
    try:
        tolerance = float(tolerance)
    except (TypeError, ValueError):
        return {"ok": False, "error": "tolerance must be a number"}

    if tolerance <= 0:
        return {"ok": False, "error": "tolerance must be positive"}

    try:
        gtype = geojson.get("type")

        if gtype == "Feature":
            geometry = geojson.get("geometry")
            if geometry and isinstance(geometry, dict):
                simplified_geom = simplify_geometry(geometry, tolerance)
                result = dict(geojson)
                result["geometry"] = simplified_geom
                return {"ok": True, "result": result}
            return {"ok": True, "result": geojson}

        elif gtype == "FeatureCollection":
            features = geojson.get("features", [])
            simplified_features = []
            for feat in features:
                if isinstance(feat, dict) and feat.get("type") == "Feature":
                    geometry = feat.get("geometry")
                    if geometry and isinstance(geometry, dict):
                        simplified_geom = simplify_geometry(geometry, tolerance)
                        new_feat = dict(feat)
                        new_feat["geometry"] = simplified_geom
                        simplified_features.append(new_feat)
                    else:
                        simplified_features.append(feat)
                else:
                    simplified_features.append(feat)
            result = dict(geojson)
            result["features"] = simplified_features
            return {"ok": True, "result": result}

        else:
            # Geometry object
            result = simplify_geometry(geojson, tolerance)
            return {"ok": True, "result": result}

    except Exception as e:
        return {"ok": False, "error": f"GeoJSON simplification failed: {str(e)}"}
