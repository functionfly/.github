import math


# Average hexagon area in km² for each H3 resolution
H3_AVG_AREA = {
    0: 4250546.848, 1: 607220.9782, 2: 86745.85403, 3: 12392.26486,
    4: 1770.323552, 5: 252.9033645, 6: 36.1290521, 7: 5.161293360,
    8: 0.7373276, 9: 0.1053325, 10: 0.0150475, 11: 0.0021496,
    12: 0.0003071, 13: 0.0000439, 14: 0.0000063, 15: 0.0000009
}


def lat_lng_to_h3_mock(lat, lng, resolution):
    """Generate a deterministic mock H3 index for given coordinates and resolution.
    
    The actual H3 library uses complex icosahedral projection. This mock generates
    a plausible-looking H3 index string based on the coordinates.
    """
    # H3 index format: resolution nibble + base cell + 15 resolution digits
    # Real H3 uses 64-bit integers encoded as hex strings
    
    # Normalize coordinates to [0, 1] range
    lat_norm = (lat + 90) / 180
    lng_norm = (lng + 180) / 360
    
    # Create a deterministic hash based on coordinates and resolution
    # Scale by resolution to get different cells at different resolutions
    scale = 7 ** resolution
    
    lat_cell = int(lat_norm * scale)
    lng_cell = int(lng_norm * scale * 2)
    
    # Combine into a 64-bit-like value
    combined = (lat_cell * 1000000007 + lng_cell * 998244353 + resolution * 12345) & 0xFFFFFFFFFFFFFF
    
    # Format as H3-like hex string
    # H3 format: 0x{resolution:01x}{base_cell:07b}{15 resolution digits}
    # Simplified: just create a plausible hex string
    hex_val = format(combined, '015x')
    
    # H3 index starts with resolution nibble
    h3_index = f"{resolution:01x}{hex_val}"
    
    # Ensure it looks like a valid H3 index (15 hex chars + resolution prefix)
    # Real H3 indices are 15 hex chars
    return h3_index[:15] + "ffff"


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

    resolution = event.get("resolution", 9)
    try:
        resolution = int(resolution)
    except (TypeError, ValueError):
        return {"ok": False, "error": "resolution must be an integer"}

    if not (0 <= resolution <= 15):
        return {"ok": False, "error": "resolution must be between 0 and 15"}

    try:
        h3_index = lat_lng_to_h3_mock(lat, lng, resolution)
        avg_area = H3_AVG_AREA.get(resolution, 0.1)

        return {
            "ok": True,
            "result": {
                "h3_index": h3_index,
                "resolution": resolution,
                "lat": lat,
                "lng": lng,
                "avg_area_km2": avg_area,
                "note": "Mock H3 index - install h3-py library for accurate H3 indexing"
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"H3 index generation failed: {str(e)}"}
