import math


CITY_DATABASE = [
    {"lat": 40.7128, "lng": -74.0060, "city": "New York", "formatted": "New York, NY, USA", "country": "United States", "code": "US"},
    {"lat": 34.0522, "lng": -118.2437, "city": "Los Angeles", "formatted": "Los Angeles, CA, USA", "country": "United States", "code": "US"},
    {"lat": 41.8781, "lng": -87.6298, "city": "Chicago", "formatted": "Chicago, IL, USA", "country": "United States", "code": "US"},
    {"lat": 29.7604, "lng": -95.3698, "city": "Houston", "formatted": "Houston, TX, USA", "country": "United States", "code": "US"},
    {"lat": 33.4484, "lng": -112.0740, "city": "Phoenix", "formatted": "Phoenix, AZ, USA", "country": "United States", "code": "US"},
    {"lat": 39.9526, "lng": -75.1652, "city": "Philadelphia", "formatted": "Philadelphia, PA, USA", "country": "United States", "code": "US"},
    {"lat": 29.4241, "lng": -98.4936, "city": "San Antonio", "formatted": "San Antonio, TX, USA", "country": "United States", "code": "US"},
    {"lat": 32.7157, "lng": -117.1611, "city": "San Diego", "formatted": "San Diego, CA, USA", "country": "United States", "code": "US"},
    {"lat": 32.7767, "lng": -96.7970, "city": "Dallas", "formatted": "Dallas, TX, USA", "country": "United States", "code": "US"},
    {"lat": 37.3382, "lng": -121.8863, "city": "San Jose", "formatted": "San Jose, CA, USA", "country": "United States", "code": "US"},
    {"lat": 30.2672, "lng": -97.7431, "city": "Austin", "formatted": "Austin, TX, USA", "country": "United States", "code": "US"},
    {"lat": 47.6062, "lng": -122.3321, "city": "Seattle", "formatted": "Seattle, WA, USA", "country": "United States", "code": "US"},
    {"lat": 39.7392, "lng": -104.9903, "city": "Denver", "formatted": "Denver, CO, USA", "country": "United States", "code": "US"},
    {"lat": 42.3601, "lng": -71.0589, "city": "Boston", "formatted": "Boston, MA, USA", "country": "United States", "code": "US"},
    {"lat": 25.7617, "lng": -80.1918, "city": "Miami", "formatted": "Miami, FL, USA", "country": "United States", "code": "US"},
    {"lat": 33.7490, "lng": -84.3880, "city": "Atlanta", "formatted": "Atlanta, GA, USA", "country": "United States", "code": "US"},
    {"lat": 37.7749, "lng": -122.4194, "city": "San Francisco", "formatted": "San Francisco, CA, USA", "country": "United States", "code": "US"},
    {"lat": 38.9072, "lng": -77.0369, "city": "Washington", "formatted": "Washington, DC, USA", "country": "United States", "code": "US"},
    {"lat": 51.5074, "lng": -0.1278, "city": "London", "formatted": "London, UK", "country": "United Kingdom", "code": "GB"},
    {"lat": 48.8566, "lng": 2.3522, "city": "Paris", "formatted": "Paris, France", "country": "France", "code": "FR"},
    {"lat": 35.6762, "lng": 139.6503, "city": "Tokyo", "formatted": "Tokyo, Japan", "country": "Japan", "code": "JP"},
    {"lat": 52.5200, "lng": 13.4050, "city": "Berlin", "formatted": "Berlin, Germany", "country": "Germany", "code": "DE"},
    {"lat": -33.8688, "lng": 151.2093, "city": "Sydney", "formatted": "Sydney, NSW, Australia", "country": "Australia", "code": "AU"},
    {"lat": 43.6532, "lng": -79.3832, "city": "Toronto", "formatted": "Toronto, ON, Canada", "country": "Canada", "code": "CA"},
    {"lat": 1.3521, "lng": 103.8198, "city": "Singapore", "formatted": "Singapore", "country": "Singapore", "code": "SG"},
    {"lat": 25.2048, "lng": 55.2708, "city": "Dubai", "formatted": "Dubai, UAE", "country": "United Arab Emirates", "code": "AE"},
    {"lat": 55.7558, "lng": 37.6173, "city": "Moscow", "formatted": "Moscow, Russia", "country": "Russia", "code": "RU"},
    {"lat": 39.9042, "lng": 116.4074, "city": "Beijing", "formatted": "Beijing, China", "country": "China", "code": "CN"},
    {"lat": 31.2304, "lng": 121.4737, "city": "Shanghai", "formatted": "Shanghai, China", "country": "China", "code": "CN"},
    {"lat": 22.3193, "lng": 114.1694, "city": "Hong Kong", "formatted": "Hong Kong", "country": "Hong Kong", "code": "HK"},
    {"lat": 37.5665, "lng": 126.9780, "city": "Seoul", "formatted": "Seoul, South Korea", "country": "South Korea", "code": "KR"},
    {"lat": 19.0760, "lng": 72.8777, "city": "Mumbai", "formatted": "Mumbai, India", "country": "India", "code": "IN"},
    {"lat": 28.7041, "lng": 77.1025, "city": "Delhi", "formatted": "Delhi, India", "country": "India", "code": "IN"},
    {"lat": 13.7563, "lng": 100.5018, "city": "Bangkok", "formatted": "Bangkok, Thailand", "country": "Thailand", "code": "TH"},
    {"lat": -6.2088, "lng": 106.8456, "city": "Jakarta", "formatted": "Jakarta, Indonesia", "country": "Indonesia", "code": "ID"},
    {"lat": 3.1390, "lng": 101.6869, "city": "Kuala Lumpur", "formatted": "Kuala Lumpur, Malaysia", "country": "Malaysia", "code": "MY"},
    {"lat": 14.5995, "lng": 120.9842, "city": "Manila", "formatted": "Manila, Philippines", "country": "Philippines", "code": "PH"},
    {"lat": 30.0444, "lng": 31.2357, "city": "Cairo", "formatted": "Cairo, Egypt", "country": "Egypt", "code": "EG"},
    {"lat": 6.5244, "lng": 3.3792, "city": "Lagos", "formatted": "Lagos, Nigeria", "country": "Nigeria", "code": "NG"},
    {"lat": -1.2921, "lng": 36.8219, "city": "Nairobi", "formatted": "Nairobi, Kenya", "country": "Kenya", "code": "KE"},
    {"lat": -26.2041, "lng": 28.0473, "city": "Johannesburg", "formatted": "Johannesburg, South Africa", "country": "South Africa", "code": "ZA"},
    {"lat": -33.9249, "lng": 18.4241, "city": "Cape Town", "formatted": "Cape Town, South Africa", "country": "South Africa", "code": "ZA"},
    {"lat": -37.8136, "lng": 144.9631, "city": "Melbourne", "formatted": "Melbourne, VIC, Australia", "country": "Australia", "code": "AU"},
    {"lat": -27.4698, "lng": 153.0251, "city": "Brisbane", "formatted": "Brisbane, QLD, Australia", "country": "Australia", "code": "AU"},
    {"lat": -23.5505, "lng": -46.6333, "city": "Sao Paulo", "formatted": "São Paulo, Brazil", "country": "Brazil", "code": "BR"},
    {"lat": -22.9068, "lng": -43.1729, "city": "Rio de Janeiro", "formatted": "Rio de Janeiro, Brazil", "country": "Brazil", "code": "BR"},
    {"lat": -34.6037, "lng": -58.3816, "city": "Buenos Aires", "formatted": "Buenos Aires, Argentina", "country": "Argentina", "code": "AR"},
    {"lat": 19.4326, "lng": -99.1332, "city": "Mexico City", "formatted": "Mexico City, Mexico", "country": "Mexico", "code": "MX"},
    {"lat": 49.2827, "lng": -123.1207, "city": "Vancouver", "formatted": "Vancouver, BC, Canada", "country": "Canada", "code": "CA"},
    {"lat": 45.5017, "lng": -73.5673, "city": "Montreal", "formatted": "Montreal, QC, Canada", "country": "Canada", "code": "CA"},
    {"lat": 40.4168, "lng": -3.7038, "city": "Madrid", "formatted": "Madrid, Spain", "country": "Spain", "code": "ES"},
    {"lat": 41.3851, "lng": 2.1734, "city": "Barcelona", "formatted": "Barcelona, Spain", "country": "Spain", "code": "ES"},
    {"lat": 41.9028, "lng": 12.4964, "city": "Rome", "formatted": "Rome, Italy", "country": "Italy", "code": "IT"},
    {"lat": 45.4654, "lng": 9.1859, "city": "Milan", "formatted": "Milan, Italy", "country": "Italy", "code": "IT"},
    {"lat": 52.3676, "lng": 4.9041, "city": "Amsterdam", "formatted": "Amsterdam, Netherlands", "country": "Netherlands", "code": "NL"},
    {"lat": 48.2082, "lng": 16.3738, "city": "Vienna", "formatted": "Vienna, Austria", "country": "Austria", "code": "AT"},
    {"lat": 47.3769, "lng": 8.5417, "city": "Zurich", "formatted": "Zurich, Switzerland", "country": "Switzerland", "code": "CH"},
    {"lat": 59.3293, "lng": 18.0686, "city": "Stockholm", "formatted": "Stockholm, Sweden", "country": "Sweden", "code": "SE"},
    {"lat": 59.9139, "lng": 10.7522, "city": "Oslo", "formatted": "Oslo, Norway", "country": "Norway", "code": "NO"},
    {"lat": 55.6761, "lng": 12.5683, "city": "Copenhagen", "formatted": "Copenhagen, Denmark", "country": "Denmark", "code": "DK"},
    {"lat": 60.1699, "lng": 24.9384, "city": "Helsinki", "formatted": "Helsinki, Finland", "country": "Finland", "code": "FI"},
    {"lat": 52.2297, "lng": 21.0122, "city": "Warsaw", "formatted": "Warsaw, Poland", "country": "Poland", "code": "PL"},
    {"lat": 50.0755, "lng": 14.4378, "city": "Prague", "formatted": "Prague, Czech Republic", "country": "Czech Republic", "code": "CZ"},
    {"lat": 47.4979, "lng": 19.0402, "city": "Budapest", "formatted": "Budapest, Hungary", "country": "Hungary", "code": "HU"},
    {"lat": 41.0082, "lng": 28.9784, "city": "Istanbul", "formatted": "Istanbul, Turkey", "country": "Turkey", "code": "TR"},
    {"lat": 24.7136, "lng": 46.6753, "city": "Riyadh", "formatted": "Riyadh, Saudi Arabia", "country": "Saudi Arabia", "code": "SA"},
    {"lat": 35.6892, "lng": 51.3890, "city": "Tehran", "formatted": "Tehran, Iran", "country": "Iran", "code": "IR"},
    {"lat": 23.8103, "lng": 90.4125, "city": "Dhaka", "formatted": "Dhaka, Bangladesh", "country": "Bangladesh", "code": "BD"},
    {"lat": 24.8607, "lng": 67.0011, "city": "Karachi", "formatted": "Karachi, Pakistan", "country": "Pakistan", "code": "PK"},
    {"lat": 12.9716, "lng": 77.5946, "city": "Bangalore", "formatted": "Bangalore, India", "country": "India", "code": "IN"},
    {"lat": 25.0330, "lng": 121.5654, "city": "Taipei", "formatted": "Taipei, Taiwan", "country": "Taiwan", "code": "TW"},
    {"lat": 34.6937, "lng": 135.5023, "city": "Osaka", "formatted": "Osaka, Japan", "country": "Japan", "code": "JP"},
    {"lat": -36.8485, "lng": 174.7633, "city": "Auckland", "formatted": "Auckland, New Zealand", "country": "New Zealand", "code": "NZ"},
    {"lat": -41.2865, "lng": 174.7762, "city": "Wellington", "formatted": "Wellington, New Zealand", "country": "New Zealand", "code": "NZ"},
    {"lat": 61.2181, "lng": -149.9003, "city": "Anchorage", "formatted": "Anchorage, AK, USA", "country": "United States", "code": "US"},
    {"lat": 21.3069, "lng": -157.8583, "city": "Honolulu", "formatted": "Honolulu, HI, USA", "country": "United States", "code": "US"},
]


def haversine_distance(lat1, lng1, lat2, lng2):
    R = 6371.0
    phi1 = math.radians(lat1)
    phi2 = math.radians(lat2)
    dphi = math.radians(lat2 - lat1)
    dlambda = math.radians(lng2 - lng1)
    a = math.sin(dphi / 2) ** 2 + math.cos(phi1) * math.cos(phi2) * math.sin(dlambda / 2) ** 2
    return R * 2 * math.atan2(math.sqrt(a), math.sqrt(1 - a))


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

    if lat < -90 or lat > 90:
        return {"ok": False, "error": "lat must be between -90 and 90"}
    if lng < -180 or lng > 180:
        return {"ok": False, "error": "lng must be between -180 and 180"}

    try:
        # Find nearest city
        nearest = None
        min_dist = float("inf")
        for city in CITY_DATABASE:
            dist = haversine_distance(lat, lng, city["lat"], city["lng"])
            if dist < min_dist:
                min_dist = dist
                nearest = city

        # Confidence based on distance (closer = higher confidence)
        if min_dist < 10:
            confidence = 0.95
        elif min_dist < 50:
            confidence = 0.85
        elif min_dist < 200:
            confidence = 0.70
        elif min_dist < 500:
            confidence = 0.55
        else:
            confidence = 0.35

        return {
            "ok": True,
            "result": {
                "lat": lat,
                "lng": lng,
                "formatted_address": nearest["formatted"],
                "city": nearest["city"],
                "country": nearest["country"],
                "country_code": nearest["code"],
                "confidence": confidence,
                "distance_km": round(min_dist, 2)
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"reverse geocoding failed: {str(e)}"}
