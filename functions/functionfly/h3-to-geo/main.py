import math
import re


H3_AVG_AREA = {
    0: 4250546.848, 1: 607220.9782, 2: 86745.85403, 3: 12392.26486,
    4: 1770.323552, 5: 252.9033645, 6: 36.1290521, 7: 5.161293360,
    8: 0.7373276, 9: 0.1053325, 10: 0.0150475, 11: 0.0021496,
    12: 0.0003071, 13: 0.0000439, 14: 0.0000063, 15: 0.0000009
}


def h3_to_geo_mock(h3_index):
    """Convert mock H3 index back to approximate coordinates.
    
    Since we generated mock H3 indices, we reverse the process.
    For real H3 indices, use the h3-py library.
    """
    h3_index = h3_index.lower().strip()
    
    # Extract resolution from first character
    try:
        resolution = int(h3_index[0], 16)
    except (ValueError, IndexError):
        resolution = 9
    
    resolution = max(0, min(15, resolution))
    
    # Extract the coordinate-encoding part
    hex_part = h3_index[1:16] if len(h3_index) >= 16 else h3_index[1:]
    
    # Reverse the mock encoding
    try:
        combined = int(hex_part[:15], 16) if len(hex_part) >= 15 else int(hex_part, 16)
    except ValueError:
        combined = 0
    
    scale = 7 ** resolution
    
    # Reverse: combined = (lat_cell * 1000000007 + lng_cell * 998244353 + resolution * 12345) & mask
    # This is not perfectly reversible, so we use an approximation
    # For a mock, we'll derive approximate coordinates from the hash
    
    # Use the hash to generate coordinates in a deterministic way
    lat_approx = ((combined % 18000) - 9000) / 100.0
    lng_approx = (((combined >> 8) % 36000) - 18000) / 100.0
    
    # Clamp to valid ranges
    lat_approx = max(-90, min(90, lat_approx))
    lng_approx = max(-180, min(180, lng_approx))
    
    return lat_approx, lng_approx, resolution


def generate_hexagon_boundary(lat, lng, resolution):
    """Generate approximate hexagon boundary for given center and resolution."""
    # Approximate hexagon radius based on resolution
    area_km2 = H3_AVG_AREA.get(resolution, 0.1)
    # Area of regular hexagon = (3√3/2) * r²
    # r = sqrt(area / (3√3/2))
    r_km = math.sqrt(area_km2 / (3 * math.sqrt(3) / 2))
    
    # Convert km to degrees (approximate)
    r_lat = r_km / 111.0
    r_lng = r_km / (111.0 * math.cos(math.radians(lat)))
    
    # Generate 6 vertices of hexagon
    boundary = []
    for i in range(6):
        angle = math.radians(60 * i + 30)  # Flat-top hexagon
        v_lat = lat + r_lat * math.sin(angle)
        v_lng = lng + r_lng * math.cos(angle)
        boundary.append([round(v_lat, 6), round(v_lng, 6)])
    
    return boundary


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}

    h3_index = event.get("h3_index")
    if h3_index is None:
        return {"ok": False, "error": "h3_index is required"}

    if not isinstance(h3_index, str):
        return {"ok": False, "error": "h3_index must be a string"}

    h3_index = h3_index.strip().lower()
    if not h3_index:
        return {"ok": False, "error": "h3_index cannot be empty"}

    # Validate hex string
    if not re.match(r'^[0-9a-f]+$', h3_index):
        return {"ok": False, "error": "h3_index must be a hexadecimal string"}

    if len(h3_index) < 1 or len(h3_index) > 16:
        return {"ok": False, "error": "h3_index must be 1-16 hex characters"}

    try:
        lat, lng, resolution = h3_to_geo_mock(h3_index)
        boundary = generate_hexagon_boundary(lat, lng, resolution)

        return {
            "ok": True,
            "result": {
                "h3_index": h3_index,
                "resolution": resolution,
                "lat": round(lat, 6),
                "lng": round(lng, 6),
                "boundary": boundary,
                "avg_area_km2": H3_AVG_AREA.get(resolution, 0.1),
                "note": "Mock H3 conversion - install h3-py library for accurate H3 operations"
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"H3 to geo conversion failed: {str(e)}"}
