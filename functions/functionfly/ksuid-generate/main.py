import os
import time
import base64

# KSUID: K-Sortable Unique Identifier
# 20 bytes: 4 bytes timestamp (seconds since 2014-05-13) + 16 bytes random payload

KSUID_EPOCH = 1400000000  # 2014-05-13 00:00:00 UTC

def generate_ksuid():
    """Generate a KSUID"""
    timestamp = int(time.time()) - KSUID_EPOCH
    timestamp_bytes = timestamp.to_bytes(4, 'big')
    payload = os.urandom(16)
    ksuid_bytes = timestamp_bytes + payload
    # Encode to base62 (KSUID uses base62 encoding)
    alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
    num = int.from_bytes(ksuid_bytes, 'big')
    if num == 0:
        return alphabet[0] * 27
    result = []
    while num > 0:
        num, remainder = divmod(num, 62)
        result.append(alphabet[remainder])
    # Pad to 27 characters
    while len(result) < 27:
        result.append(alphabet[0])
    return ''.join(reversed(result))

def handler(event):
    try:
        ksuid = generate_ksuid()
        return {"ok": True, "ksuid": ksuid}
    except Exception as e:
        return {"ok": False, "error": str(e)}
