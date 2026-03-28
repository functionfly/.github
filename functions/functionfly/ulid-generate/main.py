import os
import time

# ULID: Universally Unique Lexicographically Sortable Identifier
# 26 characters: 10 characters timestamp (milliseconds) + 16 characters random

ULID_ALPHABET = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

def generate_ulid():
    """Generate a ULID"""
    timestamp = int(time.time() * 1000)
    # Encode timestamp (10 characters)
    timestamp_chars = []
    for _ in range(10):
        timestamp_chars.append(ULID_ALPHABET[timestamp % 32])
        timestamp //= 32
    timestamp_chars.reverse()
    # Encode random (16 characters)
    random_bytes = os.urandom(10)
    random_chars = []
    for byte in random_bytes:
        random_chars.append(ULID_ALPHABET[byte % 32])
    return ''.join(timestamp_chars) + ''.join(random_chars)

def handler(event):
    try:
        ulid = generate_ulid()
        return {"ok": True, "ulid": ulid}
    except Exception as e:
        return {"ok": False, "error": str(e)}
