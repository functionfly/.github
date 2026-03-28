import math


def haversine_km(lat1, lng1, lat2, lng2):
    R = 6371.0
    phi1 = math.radians(lat1)
    phi2 = math.radians(lat2)
    dphi = math.radians(lat2 - lat1)
    dlambda = math.radians(lng2 - lng1)
    a = math.sin(dphi / 2) ** 2 + math.cos(phi1) * math.cos(phi2) * math.sin(dlambda / 2) ** 2
    return R * 2 * math.atan2(math.sqrt(a), math.sqrt(1 - a))


def point_along(parsed, target_km, segment_lengths, total_km):
    """Get point at target_km along the line."""
    if target_km <= 0:
        return list(parsed[0])
    if target_km >= total_km:
        return list(parsed[-1])

    traveled = 0.0
    for i, seg_len in enumerate(segment_lengths):
        if traveled + seg_len >= target_km:
            remaining = target_km - traveled
            fraction = remaining / seg_len if seg_len > 0 else 0
            lng = parsed[i][0] + (parsed[i+1][0] - parsed[i][0]) * fraction
            lat = parsed[i][1] + (parsed[i+1][1] - parsed[i][1]) * fraction
            return [round(lng, 6), round(lat, 6)]
        traveled += seg_len
    return list(parsed[-1])


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}

    coordinates = event.get("coordinates")
    start_distance = event.get("start_distance")
    stop_distance = event.get("stop_distance")

    if coordinates is None:
        return {"ok": False, "error": "coordinates is required"}
    if start_distance is None:
        return {"ok": False, "error": "start_distance is required"}
    if stop_distance is None:
        return {"ok": False, "error": "stop_distance is required"}

    if not isinstance(coordinates, list) or len(coordinates) < 2:
        return {"ok": False, "error": "coordinates must be an array of at least 2 points"}

    try:
        start_km = float(start_distance)
        stop_km = float(stop_distance)
    except (TypeError, ValueError):
        return {"ok": False, "error": "start_distance and stop_distance must be numbers"}

    unit = event.get("unit", "km")
    valid_units = ["km", "miles", "meters"]
    if unit not in valid_units:
        return {"ok": False, "error": f"unit must be one of: {', '.join(valid_units)}"}

    # Convert to km
    if unit == "miles":
        start_km *= 1.60934
        stop_km *= 1.60934
    elif unit == "meters":
        start_km /= 1000.0
        stop_km /= 1000.0

    if start_km > stop_km:
        start_km, stop_km = stop_km, start_km

    try:
        parsed = []
        for i, coord in enumerate(coordinates):
            if not isinstance(coord, (list, tuple)) or len(coord) < 2:
                return {"ok": False, "error": f"coordinate at index {i} must be [lng, lat]"}
            lng = float(coord[0])
            lat = float(coord[1])
            parsed.append((lng, lat))

        # Calculate segment lengths
        segment_lengths = []
        total_km = 0.0
        for i in range(len(parsed) - 1):
            seg_len = haversine_km(parsed[i][1], parsed[i][0], parsed[i+1][1], parsed[i+1][0])
            segment_lengths.append(seg_len)
            total_km += seg_len

        # Get start and stop points
        start_pt = point_along(parsed, start_km, segment_lengths, total_km)
        stop_pt = point_along(parsed, stop_km, segment_lengths, total_km)

        # Build sliced line
        sliced = [start_pt]
        traveled = 0.0
        for i, seg_len in enumerate(segment_lengths):
            if traveled + seg_len > start_km and traveled < stop_km:
                # This segment is (partially) in range
                if traveled + seg_len <= stop_km:
                    # Add the end vertex of this segment
                    sliced.append(list(parsed[i+1]))
            traveled += seg_len

        if sliced[-1] != stop_pt:
            sliced.append(stop_pt)

        # Calculate length of sliced segment
        slice_len = haversine_km(
            start_pt[1], start_pt[0],
            stop_pt[1], stop_pt[0]
        ) if len(sliced) == 2 else sum(
            haversine_km(sliced[i][1], sliced[i][0], sliced[i+1][1], sliced[i+1][0])
            for i in range(len(sliced) - 1)
        )

        return {
            "ok": True,
            "result": {
                "coordinates": sliced,
                "length_km": round(slice_len, 4),
                "start_distance": start_distance,
                "stop_distance": stop_distance,
                "unit": unit
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"line slice failed: {str(e)}"}
