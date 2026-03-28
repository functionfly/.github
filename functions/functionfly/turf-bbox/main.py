def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}

    coordinates = event.get("coordinates")
    if coordinates is None:
        return {"ok": False, "error": "coordinates is required"}

    if not isinstance(coordinates, list):
        return {"ok": False, "error": "coordinates must be an array"}

    if len(coordinates) < 1:
        return {"ok": False, "error": "coordinates must contain at least one point"}

    try:
        lngs = []
        lats = []
        for i, coord in enumerate(coordinates):
            if not isinstance(coord, (list, tuple)) or len(coord) < 2:
                return {"ok": False, "error": f"coordinate at index {i} must be [lng, lat]"}
            lng = float(coord[0])
            lat = float(coord[1])
            if not (-90 <= lat <= 90):
                return {"ok": False, "error": f"lat at index {i} must be between -90 and 90"}
            if not (-180 <= lng <= 180):
                return {"ok": False, "error": f"lng at index {i} must be between -180 and 180"}
            lngs.append(lng)
            lats.append(lat)

        min_lng = min(lngs)
        min_lat = min(lats)
        max_lng = max(lngs)
        max_lat = max(lats)

        return {
            "ok": True,
            "result": {
                "bbox": [min_lng, min_lat, max_lng, max_lat],
                "min_lng": min_lng,
                "min_lat": min_lat,
                "max_lng": max_lng,
                "max_lat": max_lat
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"bbox calculation failed: {str(e)}"}
