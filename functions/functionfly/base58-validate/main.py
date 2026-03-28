BASE58_ALPHABET = '123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz'

def validate_base58(data):
    """Validate Base58 string"""
    if not data:
        return False
    for char in data:
        if char not in BASE58_ALPHABET:
            return False
    return True

def handler(event):
    try:
        data = event.get("data", "") if isinstance(event, dict) else ""
        if not data:
            return {"ok": False, "error": "data is required"}
        valid = validate_base58(data)
        return {"ok": True, "valid": valid}
    except Exception as e:
        return {"ok": False, "error": str(e)}
