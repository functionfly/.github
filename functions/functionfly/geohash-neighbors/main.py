BASE32 = "0123456789bcdefghjkmnpqrstuvwxyz"
BASE32_MAP = {c: i for i, c in enumerate(BASE32)}

# Neighbor lookup tables
NEIGHBOR = {
    "right":  {"even": "bc01fg45telegramhijklmnopqrstuvwx", "odd": "p0r21436x8zb9dcf5h7kjnmqesgutwvy"},
    "left":   {"even": "238967debc01telegramhijklmnopqrstuvwx", "odd": "14365h7k9dcfesgujnmqp0r2twvyx8zb"},
    "top":    {"even": "p0r21436x8zb9dcf5h7kjnmqesgutwvy", "odd": "bc01fg45telegramhijklmnopqrstuvwx"},
    "bottom": {"even": "14365h7k9dcfesgujnmqp0r2twvyx8zb", "odd": "238967debc01telegramhijklmnopqrstuvwx"},
}

BORDER = {
    "right":  {"even": "bcfguvyz", "odd": "prxz"},
    "left":   {"even": "0145hjnp", "odd": "028b"},
    "top":    {"even": "prxz", "odd": "bcfguvyz"},
    "bottom": {"even": "028b", "odd": "0145hjnp"},
}

# Simplified neighbor tables (standard geohash neighbor algorithm)
NEIGHBORS = {
    "n": {"even": "p0r21436x8zb9dcf5h7kjnmqesgutwvy", "odd": "bc01fg45238967deuvhjyznpkmstqrwx"},
    "s": {"even": "14365h7k9dcfesgujnmqp0r2twvyx8zb", "odd": "238967debc01fg45telegramhijklmnopqrstuvwx"},
    "e": {"even": "bc01fg45telegramhijklmnopqrstuvwx", "odd": "p0r21436x8zb9dcf5h7kjnmqesgutwvy"},
    "w": {"even": "238967debc01telegramhijklmnopqrstuvwx", "odd": "14365h7k9dcfesgujnmqp0r2twvyx8zb"},
}

BORDERS = {
    "n": {"even": "prxz", "odd": "bcfguvyz"},
    "s": {"even": "028b", "odd": "0145hjnp"},
    "e": {"even": "bcfguvyz", "odd": "prxz"},
    "w": {"even": "0145hjnp", "odd": "028b"},
}


def decode_to_range(geohash):
    """Decode geohash to lat/lng ranges."""
    lat_range = [-90.0, 90.0]
    lng_range = [-180.0, 180.0]
    even = True

    for char in geohash.lower():
        if char not in BASE32_MAP:
            raise ValueError(f"invalid geohash character: '{char}'")
        bits = BASE32_MAP[char]
        for i in range(4, -1, -1):
            bit = (bits >> i) & 1
            if even:
                mid = (lng_range[0] + lng_range[1]) / 2
                if bit:
                    lng_range[0] = mid
                else:
                    lng_range[1] = mid
            else:
                mid = (lat_range[0] + lat_range[1]) / 2
                if bit:
                    lat_range[0] = mid
                else:
                    lat_range[1] = mid
            even = not even

    return lat_range, lng_range


def encode(lat, lng, precision):
    """Encode lat/lng to geohash."""
    lat_range = [-90.0, 90.0]
    lng_range = [-180.0, 180.0]
    geohash = []
    bits = 0
    bit_count = 0
    even = True

    while len(geohash) < precision:
        if even:
            mid = (lng_range[0] + lng_range[1]) / 2
            if lng >= mid:
                bits = (bits << 1) | 1
                lng_range[0] = mid
            else:
                bits = bits << 1
                lng_range[1] = mid
        else:
            mid = (lat_range[0] + lat_range[1]) / 2
            if lat >= mid:
                bits = (bits << 1) | 1
                lat_range[0] = mid
            else:
                bits = bits << 1
                lat_range[1] = mid
        even = not even
        bit_count += 1
        if bit_count == 5:
            geohash.append(BASE32[bits])
            bits = 0
            bit_count = 0

    return "".join(geohash)


def get_neighbor(geohash, direction):
    """Get neighboring geohash in given direction."""
    geohash = geohash.lower()
    last_char = geohash[-1]
    parent = geohash[:-1]
    precision = len(geohash)

    typ = "odd" if precision % 2 else "even"

    # Decode center and move in direction
    lat_range, lng_range = decode_to_range(geohash)
    lat = (lat_range[0] + lat_range[1]) / 2
    lng = (lng_range[0] + lng_range[1]) / 2
    lat_err = (lat_range[1] - lat_range[0]) / 2
    lng_err = (lng_range[1] - lng_range[0]) / 2

    if direction == "n":
        lat = min(90, lat + lat_err * 2)
    elif direction == "s":
        lat = max(-90, lat - lat_err * 2)
    elif direction == "e":
        lng = min(180, lng + lng_err * 2)
    elif direction == "w":
        lng = max(-180, lng - lng_err * 2)
    elif direction == "ne":
        lat = min(90, lat + lat_err * 2)
        lng = min(180, lng + lng_err * 2)
    elif direction == "nw":
        lat = min(90, lat + lat_err * 2)
        lng = max(-180, lng - lng_err * 2)
    elif direction == "se":
        lat = max(-90, lat - lat_err * 2)
        lng = min(180, lng + lng_err * 2)
    elif direction == "sw":
        lat = max(-90, lat - lat_err * 2)
        lng = max(-180, lng - lng_err * 2)

    return encode(lat, lng, precision)


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}

    geohash = event.get("geohash")
    if geohash is None:
        return {"ok": False, "error": "geohash is required"}

    if not isinstance(geohash, str):
        return {"ok": False, "error": "geohash must be a string"}

    geohash = geohash.strip().lower()
    if not geohash:
        return {"ok": False, "error": "geohash cannot be empty"}

    if len(geohash) > 12:
        return {"ok": False, "error": "geohash length cannot exceed 12"}

    # Validate characters
    for c in geohash:
        if c not in BASE32_MAP:
            return {"ok": False, "error": f"invalid geohash character: '{c}'"}

    try:
        neighbors = {}
        for direction in ["n", "ne", "e", "se", "s", "sw", "w", "nw"]:
            neighbors[direction] = get_neighbor(geohash, direction)

        all_hashes = [geohash] + list(neighbors.values())

        return {
            "ok": True,
            "result": {
                "geohash": geohash,
                "neighbors": neighbors,
                "all": all_hashes
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"geohash neighbors failed: {str(e)}"}
