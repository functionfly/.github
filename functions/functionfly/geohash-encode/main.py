BASE32 = "0123456789bcdefghjkmnpqrstuvwxyz"

# Approximate error in km for each precision level
PRECISION_ERROR = {
    1: 2500, 2: 630, 3: 78, 4: 20, 5: 2.4,
    6: 0.61, 7: 0.076, 8: 0.019, 9: 0.0024,
    10: 0.00060, 11: 0.000074, 12: 0.000019
}


def encode(lat, lng, precision=9):
    """Encode lat/lng to geohash."""
    lat_range = [-90.0, 90.0]
    lng_range = [-180.0, 180.0]

    geohash = []
    bits = 0
    bit_count = 0
    even = True

    while len(geohash) < precision:
        if even:
            mid = (lng_range[0] + lng_range[1]) / 2
            if lng >= mid:
                bits = (bits << 1) | 1
                lng_range[0] = mid
            else:
                bits = bits << 1
                lng_range[1] = mid
        else:
            mid = (lat_range[0] + lat_range[1]) / 2
            if lat >= mid:
                bits = (bits << 1) | 1
                lat_range[0] = mid
            else:
                bits = bits << 1
                lat_range[1] = mid

        even = not even
        bit_count += 1

        if bit_count == 5:
            geohash.append(BASE32[bits])
            bits = 0
            bit_count = 0

    return "".join(geohash)


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}

    lat = event.get("lat")
    lng = event.get("lng")

    if lat is None:
        return {"ok": False, "error": "lat is required"}
    if lng is None:
        return {"ok": False, "error": "lng is required"}

    try:
        lat = float(lat)
        lng = float(lng)
    except (TypeError, ValueError):
        return {"ok": False, "error": "lat and lng must be numbers"}

    if not (-90 <= lat <= 90):
        return {"ok": False, "error": "lat must be between -90 and 90"}
    if not (-180 <= lng <= 180):
        return {"ok": False, "error": "lng must be between -180 and 180"}

    precision = event.get("precision", 9)
    try:
        precision = int(precision)
    except (TypeError, ValueError):
        return {"ok": False, "error": "precision must be an integer"}

    if not (1 <= precision <= 12):
        return {"ok": False, "error": "precision must be between 1 and 12"}

    try:
        geohash = encode(lat, lng, precision)
        error_km = PRECISION_ERROR.get(precision, 0.001)

        return {
            "ok": True,
            "result": {
                "geohash": geohash,
                "precision": precision,
                "lat": lat,
                "lng": lng,
                "error_km": error_km
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"geohash encoding failed: {str(e)}"}
