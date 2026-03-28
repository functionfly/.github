import hashlib

# Hashids: Generate short, unique, non-sequential IDs from numbers
# Simplified implementation

HASHIDS_ALPHABET = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

def consistent_shuffle(alphabet, salt):
    """Consistently shuffle alphabet based on salt"""
    if not salt:
        return alphabet
    alphabet = list(alphabet)
    salt_hash = hashlib.md5(salt.encode()).hexdigest()
    salt_int = int(salt_hash[:8], 16)
    for i in range(len(alphabet) - 1, 0, -1):
        salt_int = (salt_int * 2147483647 + ord(salt_hash[i % len(salt_hash)])) % 2147483647
        j = salt_int % (i + 1)
        alphabet[i], alphabet[j] = alphabet[j], alphabet[i]
    return ''.join(alphabet)

def decode_hashid(hashid, salt=""):
    """Decode hashid to numbers"""
    if not hashid:
        return []
    # Shuffle alphabet based on salt
    alphabet = consistent_shuffle(HASHIDS_ALPHABET, salt)
    # Decode numbers
    numbers = []
    parts = hashid.split('-')
    for part in parts:
        num = 0
        for char in part:
            if char not in alphabet:
                return []
            num = num * len(alphabet) + alphabet.index(char)
        numbers.append(num)
    return numbers

def handler(event):
    try:
        hashid = event.get("hashid", "") if isinstance(event, dict) else ""
        salt = event.get("salt", "") if isinstance(event, dict) else ""
        if not hashid:
            return {"ok": False, "error": "hashid is required"}
        numbers = decode_hashid(hashid, salt)
        return {"ok": True, "numbers": numbers}
    except Exception as e:
        return {"ok": False, "error": str(e)}
