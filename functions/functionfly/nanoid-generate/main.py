import os

# NanoID: A tiny, secure, URL-friendly, unique string ID generator
# Default alphabet: A-Za-z0-9_-
NANOID_ALPHABET = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_-"

def generate_nanoid(size=21):
    """Generate a NanoID"""
    random_bytes = os.urandom(size)
    nanoid = []
    for byte in random_bytes:
        nanoid.append(NANOID_ALPHABET[byte % 64])
    return ''.join(nanoid)

def handler(event):
    try:
        size = event.get("size", 21) if isinstance(event, dict) else 21
        if size < 1:
            return {"ok": False, "error": "size must be at least 1"}
        nanoid = generate_nanoid(size)
        return {"ok": True, "nanoid": nanoid}
    except Exception as e:
        return {"ok": False, "error": str(e)}
