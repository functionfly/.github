BASE58_ALPHABET = '123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz'

def base58_decode(encoded: str) -> str:
    """Decode Base58 string"""
    if not encoded:
        return ""
    # Decode from Base58
    num = 0
    for char in encoded:
        num = num * 58 + BASE58_ALPHABET.index(char)
    # Convert integer to bytes
    if num == 0:
        return ""
    result = []
    while num > 0:
        num, remainder = divmod(num, 256)
        result.append(remainder)
    # Handle leading zeros
    for char in encoded:
        if char == '1':
            result.append(0)
        else:
            break
    return bytes(reversed(result)).decode('utf-8')

def handler(event):
    try:
        encoded = event.get("encoded", "") if isinstance(event, dict) else ""
        if not encoded:
            return {"ok": False, "error": "encoded is required"}
        data = base58_decode(encoded)
        return {"ok": True, "data": data}
    except Exception as e:
        return {"ok": False, "error": str(e)}
