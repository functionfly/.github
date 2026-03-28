import math


def bearing(lat1, lng1, lat2, lng2):
    phi1 = math.radians(lat1)
    phi2 = math.radians(lat2)
    dlambda = math.radians(lng2 - lng1)
    x = math.sin(dlambda) * math.cos(phi2)
    y = math.cos(phi1) * math.sin(phi2) - math.sin(phi1) * math.cos(phi2) * math.cos(dlambda)
    theta = math.atan2(x, y)
    return (math.degrees(theta) + 360) % 360


def compass_direction(bearing_deg):
    directions = ["N", "NNE", "NE", "ENE", "E", "ESE", "SE", "SSE",
                  "S", "SSW", "SW", "WSW", "W", "WNW", "NW", "NNW"]
    idx = round(bearing_deg / 22.5) % 16
    return directions[idx]


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}

    for field in ["lat1", "lng1", "lat2", "lng2"]:
        if event.get(field) is None:
            return {"ok": False, "error": f"{field} is required"}

    try:
        lat1 = float(event["lat1"])
        lng1 = float(event["lng1"])
        lat2 = float(event["lat2"])
        lng2 = float(event["lng2"])
    except (TypeError, ValueError):
        return {"ok": False, "error": "lat1, lng1, lat2, lng2 must be numbers"}

    if not (-90 <= lat1 <= 90):
        return {"ok": False, "error": "lat1 must be between -90 and 90"}
    if not (-180 <= lng1 <= 180):
        return {"ok": False, "error": "lng1 must be between -180 and 180"}
    if not (-90 <= lat2 <= 90):
        return {"ok": False, "error": "lat2 must be between -90 and 90"}
    if not (-180 <= lng2 <= 180):
        return {"ok": False, "error": "lng2 must be between -180 and 180"}

    try:
        initial = bearing(lat1, lng1, lat2, lng2)
        # Final bearing is reverse bearing + 180
        final = (bearing(lat2, lng2, lat1, lng1) + 180) % 360

        return {
            "ok": True,
            "result": {
                "initial_bearing": round(initial, 4),
                "final_bearing": round(final, 4),
                "compass_direction": compass_direction(initial)
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"bearing calculation failed: {str(e)}"}
