import base64

# KSUID: K-Sortable Unique Identifier
# 20 bytes: 4 bytes timestamp (seconds since 2014-05-13) + 16 bytes random payload

KSUID_EPOCH = 1400000000  # 2014-05-13 00:00:00 UTC
KSUID_ALPHABET = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

def decode_base62(s):
    """Decode base62 string to integer"""
    num = 0
    for char in s:
        num = num * 62 + KSUID_ALPHABET.index(char)
    return num

def parse_ksuid(ksuid):
    """Parse a KSUID string"""
    if len(ksuid) != 27:
        raise ValueError("KSUID must be 27 characters")
    num = decode_base62(ksuid)
    ksuid_bytes = num.to_bytes(20, 'big')
    timestamp = int.from_bytes(ksuid_bytes[:4], 'big') + KSUID_EPOCH
    payload = base64.b64encode(ksuid_bytes[4:]).decode('utf-8')
    return timestamp, payload

def handler(event):
    try:
        ksuid = event.get("ksuid", "") if isinstance(event, dict) else ""
        if not ksuid:
            return {"ok": False, "error": "ksuid is required"}
        timestamp, payload = parse_ksuid(ksuid)
        return {"ok": True, "timestamp": timestamp, "payload": payload}
    except Exception as e:
        return {"ok": False, "error": str(e)}
