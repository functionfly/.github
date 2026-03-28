import math


# Approximate S2 cell area in km² for each level
S2_AREA = {
    0: 85011012.19, 1: 21252753.05, 2: 5313188.26, 3: 1328297.07,
    4: 332074.27, 5: 83018.57, 6: 20754.64, 7: 5188.66,
    8: 1297.17, 9: 324.29, 10: 81.07, 11: 20.27,
    12: 5.07, 13: 1.27, 14: 0.32, 15: 0.079,
    16: 0.020, 17: 0.0049, 18: 0.0012, 19: 0.00031,
    20: 0.000077, 21: 0.000019, 22: 0.0000048, 23: 0.0000012,
    24: 0.00000030, 25: 0.000000075, 26: 0.000000019, 27: 0.0000000047,
    28: 0.0000000012, 29: 0.00000000029, 30: 0.000000000073
}


def lat_lng_to_s2_mock(lat, lng, level):
    """Generate a deterministic mock S2 cell ID.
    
    Real S2 uses face/i/j coordinates on a cube projection.
    This mock generates plausible-looking S2 cell IDs.
    """
    # Normalize coordinates
    lat_norm = (lat + 90) / 180
    lng_norm = (lng + 180) / 360
    
    # S2 cells are 64-bit integers
    # The top 3 bits encode the face (0-5)
    # Remaining bits encode position within face
    
    # Determine face (simplified - real S2 uses cube projection)
    # Face 0: front (lng ~0), Face 1: right (lng ~90), Face 2: back (lng ~180)
    # Face 3: left (lng ~-90), Face 4: top (lat ~90), Face 5: bottom (lat ~-90)
    if lat > 45:
        face = 4
    elif lat < -45:
        face = 5
    elif -45 <= lng <= 45:
        face = 0
    elif 45 < lng <= 135:
        face = 1
    elif lng > 135 or lng < -135:
        face = 2
    else:
        face = 3
    
    # Scale by level
    scale = 2 ** (30 - level)
    
    # Position within face
    i = int(lat_norm * (2**30))
    j = int(lng_norm * (2**30))
    
    # Combine into 64-bit cell ID
    # S2 cell ID format: face(3 bits) + level bits + trailing 1 bit
    cell_id = (face << 61) | (i << 31) | j
    
    # Apply level mask
    level_bit = 1 << (2 * (30 - level))
    cell_id = (cell_id & ~(level_bit - 1)) | level_bit
    
    # Ensure positive 64-bit value
    cell_id = cell_id & 0xFFFFFFFFFFFFFFFF
    
    return cell_id


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

    level = event.get("level", 13)
    try:
        level = int(level)
    except (TypeError, ValueError):
        return {"ok": False, "error": "level must be an integer"}

    if not (0 <= level <= 30):
        return {"ok": False, "error": "level must be between 0 and 30"}

    try:
        cell_id = lat_lng_to_s2_mock(lat, lng, level)
        cell_token = format(cell_id, '016x').rstrip('0') or '0'
        area_km2 = S2_AREA.get(level, 1.0)

        return {
            "ok": True,
            "result": {
                "cell_id": str(cell_id),
                "cell_token": cell_token,
                "level": level,
                "lat": lat,
                "lng": lng,
                "area_km2": area_km2,
                "note": "Mock S2 cell - install s2geometry or s2sphere library for accurate S2 operations"
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"S2 cell generation failed: {str(e)}"}
