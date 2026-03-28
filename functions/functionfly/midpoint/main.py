import math


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
        # Convert to radians
        phi1 = math.radians(lat1)
        phi2 = math.radians(lat2)
        lambda1 = math.radians(lng1)
        dlambda = math.radians(lng2 - lng1)

        # Midpoint formula
        Bx = math.cos(phi2) * math.cos(dlambda)
        By = math.cos(phi2) * math.sin(dlambda)

        phi_mid = math.atan2(
            math.sin(phi1) + math.sin(phi2),
            math.sqrt((math.cos(phi1) + Bx) ** 2 + By ** 2)
        )
        lambda_mid = lambda1 + math.atan2(By, math.cos(phi1) + Bx)

        mid_lat = math.degrees(phi_mid)
        mid_lng = (math.degrees(lambda_mid) + 540) % 360 - 180

        return {
            "ok": True,
            "result": {
                "lat": round(mid_lat, 6),
                "lng": round(mid_lng, 6)
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"midpoint calculation failed: {str(e)}"}
