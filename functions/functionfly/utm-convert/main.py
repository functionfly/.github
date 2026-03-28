import math


# WGS84 ellipsoid parameters
A = 6378137.0  # semi-major axis
F = 1 / 298.257223563  # flattening
B = A * (1 - F)  # semi-minor axis
E2 = 1 - (B / A) ** 2  # eccentricity squared
E = math.sqrt(E2)
K0 = 0.9996  # scale factor


def utm_zone_letter(lat):
    letters = "CDEFGHJKLMNPQRSTUVWX"
    idx = int((lat + 80) / 8)
    if 0 <= idx < len(letters):
        return letters[idx]
    return "Z"


def lat_lng_to_utm(lat, lng):
    lat_r = math.radians(lat)
    lng_r = math.radians(lng)

    zone = int((lng + 180) / 6) + 1
    lng_origin = math.radians((zone - 1) * 6 - 180 + 3)

    N = A / math.sqrt(1 - E2 * math.sin(lat_r) ** 2)
    T = math.tan(lat_r) ** 2
    C = E2 / (1 - E2) * math.cos(lat_r) ** 2
    A_ = math.cos(lat_r) * (lng_r - lng_origin)

    M = A * (
        (1 - E2 / 4 - 3 * E2**2 / 64 - 5 * E2**3 / 256) * lat_r
        - (3 * E2 / 8 + 3 * E2**2 / 32 + 45 * E2**3 / 1024) * math.sin(2 * lat_r)
        + (15 * E2**2 / 256 + 45 * E2**3 / 1024) * math.sin(4 * lat_r)
        - (35 * E2**3 / 3072) * math.sin(6 * lat_r)
    )

    easting = K0 * N * (
        A_ + (1 - T + C) * A_**3 / 6
        + (5 - 18 * T + T**2 + 72 * C - 58 * E2 / (1 - E2)) * A_**5 / 120
    ) + 500000.0

    northing = K0 * (
        M + N * math.tan(lat_r) * (
            A_**2 / 2
            + (5 - T + 9 * C + 4 * C**2) * A_**4 / 24
            + (61 - 58 * T + T**2 + 600 * C - 330 * E2 / (1 - E2)) * A_**6 / 720
        )
    )

    if lat < 0:
        northing += 10000000.0

    hemisphere = "N" if lat >= 0 else "S"
    zone_letter = utm_zone_letter(lat)

    return {
        "easting": round(easting, 2),
        "northing": round(northing, 2),
        "zone": zone,
        "hemisphere": hemisphere,
        "zone_letter": zone_letter
    }


def utm_to_lat_lng(easting, northing, zone, hemisphere):
    if hemisphere.upper() == "S":
        northing -= 10000000.0

    x = easting - 500000.0
    y = northing

    lng_origin = math.radians((zone - 1) * 6 - 180 + 3)

    M = y / K0
    mu = M / (A * (1 - E2 / 4 - 3 * E2**2 / 64 - 5 * E2**3 / 256))

    e1 = (1 - math.sqrt(1 - E2)) / (1 + math.sqrt(1 - E2))
    phi1 = mu + (3 * e1 / 2 - 27 * e1**3 / 32) * math.sin(2 * mu)
    phi1 += (21 * e1**2 / 16 - 55 * e1**4 / 32) * math.sin(4 * mu)
    phi1 += (151 * e1**3 / 96) * math.sin(6 * mu)
    phi1 += (1097 * e1**4 / 512) * math.sin(8 * mu)

    N1 = A / math.sqrt(1 - E2 * math.sin(phi1) ** 2)
    T1 = math.tan(phi1) ** 2
    C1 = E2 / (1 - E2) * math.cos(phi1) ** 2
    R1 = A * (1 - E2) / (1 - E2 * math.sin(phi1) ** 2) ** 1.5
    D = x / (N1 * K0)

    lat = phi1 - (N1 * math.tan(phi1) / R1) * (
        D**2 / 2
        - (5 + 3 * T1 + 10 * C1 - 4 * C1**2 - 9 * E2 / (1 - E2)) * D**4 / 24
        + (61 + 90 * T1 + 298 * C1 + 45 * T1**2 - 252 * E2 / (1 - E2) - 3 * C1**2) * D**6 / 720
    )

    lng = lng_origin + (
        D
        - (1 + 2 * T1 + C1) * D**3 / 6
        + (5 - 2 * C1 + 28 * T1 - 3 * C1**2 + 8 * E2 / (1 - E2) + 24 * T1**2) * D**5 / 120
    ) / math.cos(phi1)

    return {
        "lat": round(math.degrees(lat), 6),
        "lng": round(math.degrees(lng), 6)
    }


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}

    direction = event.get("direction", "to_utm")

    try:
        if direction == "to_utm":
            lat = event.get("lat")
            lng = event.get("lng")
            if lat is None:
                return {"ok": False, "error": "lat is required for to_utm conversion"}
            if lng is None:
                return {"ok": False, "error": "lng is required for to_utm conversion"}
            lat = float(lat)
            lng = float(lng)
            if not (-80 <= lat <= 84):
                return {"ok": False, "error": "lat must be between -80 and 84 for UTM"}
            if not (-180 <= lng <= 180):
                return {"ok": False, "error": "lng must be between -180 and 180"}
            result = lat_lng_to_utm(lat, lng)
            return {"ok": True, "result": result}

        elif direction == "to_latlong":
            easting = event.get("easting")
            northing = event.get("northing")
            zone = event.get("zone")
            hemisphere = event.get("hemisphere", "N")
            if easting is None:
                return {"ok": False, "error": "easting is required for to_latlong conversion"}
            if northing is None:
                return {"ok": False, "error": "northing is required for to_latlong conversion"}
            if zone is None:
                return {"ok": False, "error": "zone is required for to_latlong conversion"}
            easting = float(easting)
            northing = float(northing)
            zone = int(zone)
            if not (1 <= zone <= 60):
                return {"ok": False, "error": "zone must be between 1 and 60"}
            if hemisphere.upper() not in ("N", "S"):
                return {"ok": False, "error": "hemisphere must be 'N' or 'S'"}
            result = utm_to_lat_lng(easting, northing, zone, hemisphere)
            return {"ok": True, "result": result}

        else:
            return {"ok": False, "error": "direction must be 'to_utm' or 'to_latlong'"}

    except Exception as e:
        return {"ok": False, "error": f"UTM conversion failed: {str(e)}"}
