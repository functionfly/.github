import base64

# ULID: Universally Unique Lexicographically Sortable Identifier
# 26 characters: 10 characters timestamp (milliseconds) + 16 characters random

ULID_ALPHABET = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

def decode_crockford(s):
    """Decode Crockford base32 string to integer"""
    num = 0
    for char in s.upper():
        num = num * 32 + ULID_ALPHABET.index(char)
    return num

def parse_ulid(ulid):
    """Parse a ULID string"""
    if len(ulid) != 26:
        raise ValueError("ULID must be 26 characters")
    timestamp_part = ulid[:10]
    randomness_part = ulid[10:]
    timestamp = decode_crockford(timestamp_part)
    randomness = randomness_part
    return timestamp, randomness

def handler(event):
    try:
        ulid = event.get("ulid", "") if isinstance(event, dict) else ""
        if not ulid:
            return {"ok": False, "error": "ulid is required"}
        timestamp, randomness = parse_ulid(ulid)
        return {"ok": True, "timestamp": timestamp, "randomness": randomness}
    except Exception as e:
        return {"ok": False, "error": str(e)}
