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

def encode_numbers(numbers, salt="", min_length=0):
    """Encode numbers to hashid"""
    if not numbers:
        return ""
    # Shuffle alphabet based on salt
    alphabet = consistent_shuffle(HASHIDS_ALPHABET, salt)
    # Encode numbers
    result = []
    for num in numbers:
        if num < 0:
            return ""
        if num == 0:
            result.append(alphabet[0])
        else:
            encoded = []
            while num > 0:
                encoded.append(alphabet[num % len(alphabet)])
                num //= len(alphabet)
            result.append(''.join(reversed(encoded)))
    # Join with separator
    hashid = '-'.join(result)
    # Pad to min_length
    while len(hashid) < min_length:
        hashid = alphabet[0] + hashid
    return hashid

def handler(event):
    try:
        numbers = event.get("numbers", []) if isinstance(event, dict) else []
        salt = event.get("salt", "") if isinstance(event, dict) else ""
        min_length = event.get("min_length", 0) if isinstance(event, dict) else 0
        if not numbers:
            return {"ok": False, "error": "numbers is required"}
        if not isinstance(numbers, list):
            return {"ok": False, "error": "numbers must be a list"}
        for num in numbers:
            if not isinstance(num, int) or num < 0:
                return {"ok": False, "error": "numbers must be non-negative integers"}
        hashid = encode_numbers(numbers, salt, min_length)
        return {"ok": True, "hashid": hashid}
    except Exception as e:
        return {"ok": False, "error": str(e)}
