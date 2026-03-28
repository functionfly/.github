import hashlib

def validate_ethereum_address(address):
    """Validate Ethereum address"""
    if not address:
        return False, ""
    # Check format
    if not address.startswith('0x'):
        return False, ""
    if len(address) != 42:
        return False, ""
    # Check if hex
    try:
        int(address[2:], 16)
    except ValueError:
        return False, ""
    # Check checksum (EIP-55)
    addr_lower = address[2:].lower()
    addr_hash = hashlib.sha3_256(addr_lower.encode()).hexdigest()
    checksummed = '0x'
    for i, char in enumerate(addr_lower):
        if char in '0123456789':
            checksummed += char
        elif int(addr_hash[i], 16) >= 8:
            checksummed += char.upper()
        else:
            checksummed += char.lower()
    return True, checksummed

def handler(event):
    try:
        address = event.get("address", "") if isinstance(event, dict) else ""
        if not address:
            return {"ok": False, "error": "address is required"}
        valid, checksum = validate_ethereum_address(address)
        return {"ok": True, "valid": valid, "checksum": checksum}
    except Exception as e:
        return {"ok": False, "error": str(e)}
