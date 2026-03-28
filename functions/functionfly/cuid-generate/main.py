import os
import time
import hashlib

# CUID: Collision-resistant Unique Identifier
# Format: c + timestamp (8 chars) + counter (4 chars) + fingerprint (4 chars) + random (8 chars)

CUID_ALPHABET = "abcdefghijklmnopqrstuvwxyz0123456789"

def base36_encode(num):
    """Encode number to base36"""
    if num == 0:
        return '0'
    result = []
    while num > 0:
        num, remainder = divmod(num, 36)
        result.append(CUID_ALPHABET[remainder])
    return ''.join(reversed(result))

def generate_cuid():
    """Generate a CUID"""
    timestamp = int(time.time() * 1000)
    # Counter (simplified)
    counter = int(time.time() * 1000) % 10000
    # Fingerprint (simplified - hash of hostname + pid)
    fingerprint = hashlib.md5(f"{os.getpid()}".encode()).hexdigest()[:4]
    # Random
    random_part = os.urandom(4).hex()[:8]
    # Build CUID
    cuid = 'c' + base36_encode(timestamp)[:8] + base36_encode(counter)[:4] + fingerprint + random_part
    return cuid

def handler(event):
    try:
        cuid = generate_cuid()
        return {"ok": True, "cuid": cuid}
    except Exception as e:
        return {"ok": False, "error": str(e)}
