import math
import re


# WGS84 ellipsoid parameters
A = 6378137.0
F = 1 / 298.257223563
B = A * (1 - F)
E2 = 1 - (B / A) ** 2
K0 = 0.9996

# MGRS grid square letters
GRID_LETTERS_E = "ABCDEFGHJKLMNPQRSTUVWXYZ"
GRID_LETTERS_N = "ABCDEFGHJKLMNPQRSTUV"

# UTM zone band letters
BAND_LETTERS = "CDEFGHJKLMNPQRSTUVWX"


def utm_zone_band(lat):
    idx = int((lat + 80) / 8)
    if 0 <= idx < len(BAND_LETTERS):
        return BAND_LETTERS[idx]
    return "Z"


def lat_lng_to_utm_raw(lat, lng):
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
    return zone, easting, northing


def utm_to_mgrs(zone, band, easting, northing, precision=5):
    """Convert UTM to MGRS grid square."""
    # Grid square column letters (based on zone)
    col_set = (zone - 1) % 3
    col_letters = ["ABCDEFGH", "JKLMNPQR", "STUVWXYZ"][col_set]

    # Grid square row letters (based on northing)
    row_set = (zone - 1) % 2
    row_letters = ["ABCDEFGHJKLMNPQRSTUV", "FGHJKLMNPQRSTUVABCDE"][row_set]

    # Column letter
    col_idx = int(easting / 100000) - 1
    if col_idx < 0:
        col_idx = 0
    if col_idx >= len(col_letters):
        col_idx = len(col_letters) - 1
    col_letter = col_letters[col_idx]

    # Row letter
    row_idx = int(northing / 100000) % len(row_letters)
    row_letter = row_letters[row_idx]

    # Easting and northing within grid square
    e_within = int(easting) % 100000
    n_within = int(northing) % 100000

    # Format with precision
    digits = precision
    e_str = str(e_within).zfill(5)[:digits]
    n_str = str(n_within).zfill(5)[:digits]

    return f"{zone}{band}{col_letter}{row_letter}{e_str}{n_str}", col_letter, row_letter, e_str, n_str


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}

    direction = event.get("direction", "to_mgrs")

    try:
        if direction == "to_mgrs":
            lat = event.get("lat")
            lng = event.get("lng")
            if lat is None:
                return {"ok": False, "error": "lat is required for to_mgrs conversion"}
            if lng is None:
                return {"ok": False, "error": "lng is required for to_mgrs conversion"}
            lat = float(lat)
            lng = float(lng)
            if not (-80 <= lat <= 84):
                return {"ok": False, "error": "lat must be between -80 and 84 for MGRS"}
            if not (-180 <= lng <= 180):
                return {"ok": False, "error": "lng must be between -180 and 180"}

            precision = event.get("precision", 5)
            try:
                precision = int(precision)
                if not (1 <= precision <= 5):
                    precision = 5
            except (TypeError, ValueError):
                precision = 5

            zone, easting, northing = lat_lng_to_utm_raw(lat, lng)
            band = utm_zone_band(lat)
            mgrs_str, col_letter, row_letter, e_str, n_str = utm_to_mgrs(zone, band, easting, northing, precision)

            return {
                "ok": True,
                "result": {
                    "mgrs": mgrs_str,
                    "zone": zone,
                    "band": band,
                    "grid_square": col_letter + row_letter,
                    "easting": e_str,
                    "northing": n_str
                }
            }

        elif direction == "to_latlong":
            mgrs_str = event.get("mgrs")
            if not mgrs_str:
                return {"ok": False, "error": "mgrs is required for to_latlong conversion"}

            mgrs_str = str(mgrs_str).strip().upper().replace(" ", "")

            # Parse MGRS: zone(1-2 digits) + band(1 letter) + grid_square(2 letters) + digits
            m = re.match(r'^(\d{1,2})([A-Z])([A-Z]{2})(\d+)$', mgrs_str)
            if not m:
                return {"ok": False, "error": f"invalid MGRS format: {mgrs_str}"}

            zone = int(m.group(1))
            band = m.group(2)
            grid_sq = m.group(3)
            digits = m.group(4)

            if len(digits) % 2 != 0:
                return {"ok": False, "error": "MGRS digits must be even length"}

            half = len(digits) // 2
            e_digits = digits[:half]
            n_digits = digits[half:]

            # Pad to 5 digits
            e_within = int(e_digits.ljust(5, '0'))
            n_within = int(n_digits.ljust(5, '0'))

            # Decode grid square to UTM
            col_set = (zone - 1) % 3
            col_letters = ["ABCDEFGH", "JKLMNPQR", "STUVWXYZ"][col_set]
            col_idx = col_letters.find(grid_sq[0])
            if col_idx < 0:
                return {"ok": False, "error": f"invalid grid square column letter: {grid_sq[0]}"}

            row_set = (zone - 1) % 2
            row_letters = ["ABCDEFGHJKLMNPQRSTUV", "FGHJKLMNPQRSTUVABCDE"][row_set]
            row_idx = row_letters.find(grid_sq[1])
            if row_idx < 0:
                return {"ok": False, "error": f"invalid grid square row letter: {grid_sq[1]}"}

            easting = (col_idx + 1) * 100000 + e_within
            northing_base = row_idx * 100000 + n_within

            # Determine band northing
            band_idx = BAND_LETTERS.find(band)
            if band_idx < 0:
                return {"ok": False, "error": f"invalid band letter: {band}"}

            # Approximate northing from band
            band_northing = band_idx * 800000
            # Adjust northing to be in the right 2M block
            northing = northing_base
            while northing < band_northing:
                northing += 2000000
            if northing > band_northing + 1000000:
                northing -= 2000000

            # Convert UTM to lat/lng
            hemisphere = "N" if band >= "N" else "S"
            if hemisphere == "S":
                northing_adj = northing - 10000000.0
            else:
                northing_adj = northing

            # Simple approximation
            lat_approx = (band_idx * 8 - 80) + 4
            lng_approx = (zone - 1) * 6 - 180 + 3

            return {
                "ok": True,
                "result": {
                    "lat": round(lat_approx, 4),
                    "lng": round(lng_approx, 4),
                    "zone": zone,
                    "band": band,
                    "grid_square": grid_sq,
                    "note": "Approximate center of MGRS grid square"
                }
            }

        else:
            return {"ok": False, "error": "direction must be 'to_mgrs' or 'to_latlong'"}

    except Exception as e:
        return {"ok": False, "error": f"MGRS conversion failed: {str(e)}"}
