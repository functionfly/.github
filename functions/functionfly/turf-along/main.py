import math


def haversine_km(lat1, lng1, lat2, lng2):
    R = 6371.0
    phi1 = math.radians(lat1)
    phi2 = math.radians(lat2)
    dphi = math.radians(lat2 - lat1)
    dlambda = math.radians(lng2 - lng1)
    a = math.sin(dphi / 2) ** 2 + math.cos(phi1) * math.cos(phi2) * math.sin(dlambda / 2) ** 2
    return R * 2 * math.atan2(math.sqrt(a), math.sqrt(1 - a))


def interpolate_point(lng1, lat1, lng2, lat2, fraction):
    """Linearly interpolate between two points."""
    return lng1 + (lng2 - lng1) * fraction, lat1 + (lat2 - lat1) * fraction


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}

    coordinates = event.get("coordinates")
    distance = event.get("distance")

    if coordinates is None:
        return {"ok": False, "error": "coordinates is required"}
    if distance is None:
        return {"ok": False, "error": "distance is required"}

    if not isinstance(coordinates, list) or len(coordinates) < 2:
        return {"ok": False, "error": "coordinates must be an array of at least 2 points"}

    try:
        distance = float(distance)
    except (TypeError, ValueError):
        return {"ok": False, "error": "distance must be a number"}

    if distance < 0:
        return {"ok": False, "error": "distance must be non-negative"}

    unit = event.get("unit", "km")
    valid_units = ["km", "miles", "meters"]
    if unit not in valid_units:
        return {"ok": False, "error": f"unit must be one of: {', '.join(valid_units)}"}

    try:
        parsed = []
        for i, coord in enumerate(coordinates):
            if not isinstance(coord, (list, tuple)) or len(coord) < 2:
                return {"ok": False, "error": f"coordinate at index {i} must be [lng, lat]"}
            lng = float(coord[0])
            lat = float(coord[1])
            parsed.append((lng, lat))

        # Convert distance to km
        if unit == "miles":
            dist_km = distance * 1.60934
        elif unit == "meters":
            dist_km = distance / 1000.0
        else:
            dist_km = distance

        # Calculate total length
        total_km = 0.0
        segment_lengths = []
        for i in range(len(parsed) - 1):
            seg_len = haversine_km(parsed[i][1], parsed[i][0], parsed[i+1][1], parsed[i+1][0])
            segment_lengths.append(seg_len)
            total_km += seg_len

        # Clamp distance to line length
        if dist_km >= total_km:
            last = parsed[-1]
            return {
                "ok": True,
                "result": {
                    "lng": round(last[0], 6),
                    "lat": round(last[1], 6),
                    "distance_along": round(total_km, 4),
                    "total_length_km": round(total_km, 4)
                }
            }

        # Find the segment containing the target distance
        traveled = 0.0
        for i, seg_len in enumerate(segment_lengths):
            if traveled + seg_len >= dist_km:
                # Interpolate within this segment
                remaining = dist_km - traveled
                fraction = remaining / seg_len if seg_len > 0 else 0
                lng, lat = interpolate_point(
                    parsed[i][0], parsed[i][1],
                    parsed[i+1][0], parsed[i+1][1],
                    fraction
                )
                return {
                    "ok": True,
                    "result": {
                        "lng": round(lng, 6),
                        "lat": round(lat, 6),
                        "distance_along": round(dist_km, 4),
                        "total_length_km": round(total_km, 4)
                    }
                }
            traveled += seg_len

        # Fallback to last point
        last = parsed[-1]
        return {
            "ok": True,
            "result": {
                "lng": round(last[0], 6),
                "lat": round(last[1], 6),
                "distance_along": round(total_km, 4),
                "total_length_km": round(total_km, 4)
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"along calculation failed: {str(e)}"}
