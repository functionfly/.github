def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}

    lat = event.get("lat")
    lng = event.get("lng")

    if lat is None:
        return {"ok": False, "error": "lat is required"}
    if lng is None:
        return {"ok": False, "error": "lng is required"}

    errors = []
    lat_valid = True
    lng_valid = True

    try:
        lat = float(lat)
    except (TypeError, ValueError):
        errors.append("lat must be a number")
        lat_valid = False
        lat = None

    try:
        lng = float(lng)
    except (TypeError, ValueError):
        errors.append("lng must be a number")
        lng_valid = False
        lng = None

    if lat is not None:
        if lat < -90 or lat > 90:
            errors.append(f"lat {lat} is out of range (-90 to 90)")
            lat_valid = False
        if lat != lat:  # NaN check
            errors.append("lat is NaN")
            lat_valid = False

    if lng is not None:
        if lng < -180 or lng > 180:
            errors.append(f"lng {lng} is out of range (-180 to 180)")
            lng_valid = False
        if lng != lng:  # NaN check
            errors.append("lng is NaN")
            lng_valid = False

    valid = lat_valid and lng_valid

    result = {
        "valid": valid,
        "lat_valid": lat_valid,
        "lng_valid": lng_valid,
        "lat": lat,
        "lng": lng,
        "errors": errors
    }

    if lat is not None and lat_valid:
        result["hemisphere_lat"] = "N" if lat >= 0 else "S"
    if lng is not None and lng_valid:
        result["hemisphere_lng"] = "E" if lng >= 0 else "W"

    if lat is not None and lat_valid and lng is not None and lng_valid:
        # Classify location type
        if lat == 0 and lng == 0:
            result["note"] = "Null Island (0,0)"
        elif abs(lat) > 66.5:
            result["polar_region"] = "Arctic" if lat > 0 else "Antarctic"

    return {"ok": True, "result": result}
