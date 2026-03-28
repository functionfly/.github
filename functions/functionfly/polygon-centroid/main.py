import math


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}

    coordinates = event.get("coordinates")
    if coordinates is None:
        return {"ok": False, "error": "coordinates is required"}

    if not isinstance(coordinates, list):
        return {"ok": False, "error": "coordinates must be an array"}

    if len(coordinates) < 3:
        return {"ok": False, "error": "coordinates must contain at least 3 points to form a polygon"}

    try:
        parsed = []
        for i, coord in enumerate(coordinates):
            if not isinstance(coord, (list, tuple)) or len(coord) < 2:
                return {"ok": False, "error": f"coordinate at index {i} must be [lat, lng]"}
            lat = float(coord[0])
            lng = float(coord[1])
            if not (-90 <= lat <= 90):
                return {"ok": False, "error": f"lat at index {i} must be between -90 and 90"}
            if not (-180 <= lng <= 180):
                return {"ok": False, "error": f"lng at index {i} must be between -180 and 180"}
            parsed.append((lat, lng))

        # Use weighted centroid based on polygon area (shoelace formula)
        # For geographic coordinates, use simple average for small polygons
        # For larger polygons, use the proper centroid formula
        n = len(parsed)

        # Shoelace-based centroid
        area = 0.0
        cx = 0.0
        cy = 0.0

        for i in range(n):
            j = (i + 1) % n
            x0, y0 = parsed[i][1], parsed[i][0]  # lng, lat
            x1, y1 = parsed[j][1], parsed[j][0]

            cross = x0 * y1 - x1 * y0
            area += cross
            cx += (x0 + x1) * cross
            cy += (y0 + y1) * cross

        area /= 2.0

        if abs(area) < 1e-10:
            # Degenerate polygon, use simple average
            centroid_lat = sum(p[0] for p in parsed) / n
            centroid_lng = sum(p[1] for p in parsed) / n
        else:
            cx /= (6.0 * area)
            cy /= (6.0 * area)
            centroid_lng = cx
            centroid_lat = cy

        return {
            "ok": True,
            "result": {
                "lat": round(centroid_lat, 6),
                "lng": round(centroid_lng, 6)
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"polygon centroid calculation failed: {str(e)}"}
