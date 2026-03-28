BASE32 = "0123456789bcdefghjkmnpqrstuvwxyz"
BASE32_MAP = {c: i for i, c in enumerate(BASE32)}


def decode(geohash):
    """Decode geohash to (lat, lng, lat_err, lng_err)."""
    lat_range = [-90.0, 90.0]
    lng_range = [-180.0, 180.0]
    even = True

    for char in geohash.lower():
        if char not in BASE32_MAP:
            raise ValueError(f"invalid geohash character: '{char}'")

        bits = BASE32_MAP[char]
        for i in range(4, -1, -1):
            bit = (bits >> i) & 1
            if even:
                mid = (lng_range[0] + lng_range[1]) / 2
                if bit:
                    lng_range[0] = mid
                else:
                    lng_range[1] = mid
            else:
                mid = (lat_range[0] + lat_range[1]) / 2
                if bit:
                    lat_range[0] = mid
                else:
                    lat_range[1] = mid
            even = not even

    lat = (lat_range[0] + lat_range[1]) / 2
    lng = (lng_range[0] + lng_range[1]) / 2
    lat_err = (lat_range[1] - lat_range[0]) / 2
    lng_err = (lng_range[1] - lng_range[0]) / 2

    return lat, lng, lat_err, lng_err, lat_range, lng_range


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}

    geohash = event.get("geohash")
    if geohash is None:
        return {"ok": False, "error": "geohash is required"}

    if not isinstance(geohash, str):
        return {"ok": False, "error": "geohash must be a string"}

    geohash = geohash.strip().lower()
    if not geohash:
        return {"ok": False, "error": "geohash cannot be empty"}

    if len(geohash) > 12:
        return {"ok": False, "error": "geohash length cannot exceed 12"}

    try:
        lat, lng, lat_err, lng_err, lat_range, lng_range = decode(geohash)

        # Round to appropriate precision
        decimals = min(12, len(geohash) * 2)

        return {
            "ok": True,
            "result": {
                "lat": round(lat, decimals),
                "lng": round(lng, decimals),
                "lat_error": round(lat_err, decimals),
                "lng_error": round(lng_err, decimals),
                "bbox": {
                    "min_lat": round(lat_range[0], decimals),
                    "min_lng": round(lng_range[0], decimals),
                    "max_lat": round(lat_range[1], decimals),
                    "max_lng": round(lng_range[1], decimals)
                },
                "precision": len(geohash),
                "geohash": geohash
            }
        }

    except ValueError as e:
        return {"ok": False, "error": str(e)}
    except Exception as e:
        return {"ok": False, "error": f"geohash decoding failed: {str(e)}"}
