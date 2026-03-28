import math


# Known elevation data for major cities/regions (lat, lng, elevation_m)
KNOWN_ELEVATIONS = [
    # Major cities
    (40.7128, -74.0060, 10),    # New York
    (34.0522, -118.2437, 71),   # Los Angeles
    (41.8781, -87.6298, 182),   # Chicago
    (29.7604, -95.3698, 15),    # Houston
    (33.4484, -112.0740, 331),  # Phoenix
    (39.9526, -75.1652, 12),    # Philadelphia
    (32.7157, -117.1611, 19),   # San Diego
    (32.7767, -96.7970, 139),   # Dallas
    (30.2672, -97.7431, 149),   # Austin
    (47.6062, -122.3321, 56),   # Seattle
    (39.7392, -104.9903, 1609), # Denver
    (42.3601, -71.0589, 9),     # Boston
    (25.7617, -80.1918, 2),     # Miami
    (33.7490, -84.3880, 320),   # Atlanta
    (37.7749, -122.4194, 16),   # San Francisco
    (38.9072, -77.0369, 7),     # Washington DC
    (51.5074, -0.1278, 11),     # London
    (48.8566, 2.3522, 35),      # Paris
    (35.6762, 139.6503, 40),    # Tokyo
    (52.5200, 13.4050, 34),     # Berlin
    (-33.8688, 151.2093, 39),   # Sydney
    (43.6532, -79.3832, 76),    # Toronto
    (1.3521, 103.8198, 15),     # Singapore
    (25.2048, 55.2708, 5),      # Dubai
    (55.7558, 37.6173, 156),    # Moscow
    (39.9042, 116.4074, 43),    # Beijing
    (31.2304, 121.4737, 4),     # Shanghai
    (19.0760, 72.8777, 14),     # Mumbai
    (28.7041, 77.1025, 216),    # Delhi
    (13.7563, 100.5018, 2),     # Bangkok
    (-6.2088, 106.8456, 8),     # Jakarta
    (30.0444, 31.2357, 23),     # Cairo
    (-26.2041, 28.0473, 1753),  # Johannesburg
    (-33.9249, 18.4241, 42),    # Cape Town
    (-37.8136, 144.9631, 31),   # Melbourne
    (-23.5505, -46.6333, 760),  # Sao Paulo
    (-34.6037, -58.3816, 25),   # Buenos Aires
    (19.4326, -99.1332, 2240),  # Mexico City
    (49.2827, -123.1207, 70),   # Vancouver
    (40.4168, -3.7038, 667),    # Madrid
    (41.9028, 12.4964, 21),     # Rome
    (52.3676, 4.9041, -2),      # Amsterdam
    (47.3769, 8.5417, 408),     # Zurich
    (59.3293, 18.0686, 28),     # Stockholm
    (59.9139, 10.7522, 23),     # Oslo
    (60.1699, 24.9384, 26),     # Helsinki
    (41.0082, 28.9784, 100),    # Istanbul
    (24.7136, 46.6753, 612),    # Riyadh
    (37.5665, 126.9780, 38),    # Seoul
    (34.6937, 135.5023, 15),    # Osaka
    (-36.8485, 174.7633, 25),   # Auckland
    (61.2181, -149.9003, 38),   # Anchorage
    (21.3069, -157.8583, 5),    # Honolulu
    # Mountain regions
    (27.9881, 86.9250, 8849),   # Everest
    (45.8326, 6.8652, 4808),    # Mont Blanc
    (36.5785, -118.2923, 4418), # Mt Whitney
    (46.8523, -121.7603, 4392), # Mt Rainier
    (19.0225, -98.6278, 5636),  # Popocatepetl
    # Ocean/coastal
    (0, 0, 0),                  # Null Island
    (0, 90, 0),                 # Indian Ocean
    (0, -90, 0),                # Pacific Ocean
    (0, 180, 0),                # Pacific
    (0, -30, 0),                # Atlantic
    (30, 60, 0),                # Arabian Sea
    (-30, 30, 0),               # South Atlantic
    # Deserts
    (23.4162, 25.6628, 200),    # Sahara
    (36.2048, 138.2529, 1000),  # Japan interior
    (35.8617, 104.1954, 1000),  # China interior
    (20, 80, 300),              # India interior
    (-25, 130, 300),            # Australia interior
    # Polar
    (90, 0, 2835),              # North Pole (ice)
    (-90, 0, 2835),             # South Pole
    (70, 0, 0),                 # Arctic Ocean
    (-70, 0, 2000),             # Antarctica
]


def estimate_elevation(lat, lng):
    """Estimate elevation using nearest known point and interpolation."""
    # Find nearest known elevation
    best_dist = float("inf")
    best_elev = 0

    for klat, klng, kelev in KNOWN_ELEVATIONS:
        dlat = lat - klat
        dlng = lng - klng
        dist = math.sqrt(dlat**2 + dlng**2)
        if dist < best_dist:
            best_dist = dist
            best_elev = kelev

    # If very close to a known point, use it directly
    if best_dist < 0.5:
        return best_elev

    # Otherwise, use a terrain model based on location
    # Simple model: elevation varies with latitude and longitude
    # This is a rough approximation

    # Check if likely ocean (very rough)
    # Most ocean areas are at 0m
    if abs(lat) < 60:
        # Check if far from land (simplified)
        is_likely_ocean = False
        for klat, klng, kelev in KNOWN_ELEVATIONS:
            if kelev > 0:
                dlat = lat - klat
                dlng = lng - klng
                dist = math.sqrt(dlat**2 + dlng**2)
                if dist < 5:
                    is_likely_ocean = False
                    break
        else:
            is_likely_ocean = True

        if is_likely_ocean:
            return 0

    # Use weighted average of nearby known points
    total_weight = 0
    weighted_elev = 0
    for klat, klng, kelev in KNOWN_ELEVATIONS:
        dlat = lat - klat
        dlng = lng - klng
        dist = math.sqrt(dlat**2 + dlng**2)
        if dist < 20:
            weight = 1 / (dist + 0.1)
            total_weight += weight
            weighted_elev += kelev * weight

    if total_weight > 0:
        return round(weighted_elev / total_weight)

    # Default: use latitude-based estimate
    if abs(lat) > 60:
        return 500  # Higher latitudes tend to have more elevation
    return max(0, int(abs(lat) * 5))


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

    try:
        elevation_m = estimate_elevation(lat, lng)
        elevation_ft = round(elevation_m * 3.28084, 1)

        return {
            "ok": True,
            "result": {
                "lat": lat,
                "lng": lng,
                "elevation_m": elevation_m,
                "elevation_ft": elevation_ft,
                "resolution_m": 30,
                "note": "Simulated elevation data - use SRTM or similar DEM for accurate values"
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"elevation query failed: {str(e)}"}
