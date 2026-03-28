BASE58_ALPHABET = '123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz'

def base58_encode(data: str) -> str:
    """Encode string to Base58"""
    if not data:
        return ""
    # Convert string to bytes
    data_bytes = data.encode('utf-8')
    # Convert bytes to integer
    num = int.from_bytes(data_bytes, 'big')
    # Encode to Base58
    result = []
    while num > 0:
        num, remainder = divmod(num, 58)
        result.append(BASE58_ALPHABET[remainder])
    # Handle leading zeros
    for byte in data_bytes:
        if byte == 0:
            result.append('1')
        else:
            break
    return ''.join(reversed(result))

def handler(event):
    try:
        data = event.get("data", "") if isinstance(event, dict) else ""
        if not data:
            return {"ok": False, "error": "data is required"}
        encoded = base58_encode(data)
        return {"ok": True, "encoded": encoded}
    except Exception as e:
        return {"ok": False, "error": str(e)}
