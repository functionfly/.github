import os
import time

# TSID: Time-Sorted ID
# Similar to ULID but with different encoding
# 26 characters: 10 characters timestamp (milliseconds) + 16 characters random

TSID_ALPHABET = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

def generate_tsid():
    """Generate a TSID"""
    timestamp = int(time.time() * 1000)
    # Encode timestamp (10 characters)
    timestamp_chars = []
    for _ in range(10):
        timestamp_chars.append(TSID_ALPHABET[timestamp % 32])
        timestamp //= 32
    timestamp_chars.reverse()
    # Encode random (16 characters)
    random_bytes = os.urandom(10)
    random_chars = []
    for byte in random_bytes:
        random_chars.append(TSID_ALPHABET[byte % 32])
    return ''.join(timestamp_chars) + ''.join(random_chars)

def handler(event):
    try:
        tsid = generate_tsid()
        return {"ok": True, "tsid": tsid}
    except Exception as e:
        return {"ok": False, "error": str(e)}
