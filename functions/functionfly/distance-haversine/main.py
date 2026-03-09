import math

def handler(event):
    if isinstance(event, dict):
        lat1, lon1 = event.get("lat1"), event.get("lon1")
        lat2, lon2 = event.get("lat2"), event.get("lon2")
        R = event.get("radius_km", 6371.0)
    else:
        lat1 = lon1 = lat2 = lon2 = None
        R = 6371.0
    if lat1 is None or lon1 is None or lat2 is None or lon2 is None:
        return {"ok": False, "error": "lat1, lon1, lat2, lon2 are required"}
    try:
        lat1, lon1 = math.radians(float(lat1)), math.radians(float(lon1))
        lat2, lon2 = math.radians(float(lat2)), math.radians(float(lon2))
        R = float(R)
        dlat = lat2 - lat1
        dlon = lon2 - lon1
        a = math.sin(dlat/2)**2 + math.cos(lat1) * math.cos(lat2) * math.sin(dlon/2)**2
        c = 2 * math.asin(math.sqrt(min(1, a)))
        return {"ok": True, "distance_km": round(R * c, 6)}
    except (TypeError, ValueError) as e:
        return {"ok": False, "error": str(e)}
