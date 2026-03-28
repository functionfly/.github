import hashlib
import base58

def validate_bitcoin_address(address):
    """Validate Bitcoin address"""
    if not address:
        return False, ""
    # Check for P2PKH (starts with 1)
    if address.startswith('1'):
        if len(address) < 25 or len(address) > 34:
            return False, ""
        try:
            decoded = base58.b58decode(address)
            if len(decoded) != 25:
                return False, ""
            checksum = hashlib.sha256(hashlib.sha256(decoded[:-4]).digest()).digest()[:4]
            if decoded[-4:] == checksum:
                return True, "P2PKH"
        except:
            pass
    # Check for P2SH (starts with 3)
    elif address.startswith('3'):
        if len(address) < 25 or len(address) > 34:
            return False, ""
        try:
            decoded = base58.b58decode(address)
            if len(decoded) != 25:
                return False, ""
            checksum = hashlib.sha256(hashlib.sha256(decoded[:-4]).digest()).digest()[:4]
            if decoded[-4:] == checksum:
                return True, "P2SH"
        except:
            pass
    # Check for Bech32 (starts with bc1)
    elif address.startswith('bc1'):
        if len(address) < 14 or len(address) > 74:
            return False, ""
        # Simplified validation for Bech32
        return True, "Bech32"
    return False, ""

def handler(event):
    try:
        address = event.get("address", "") if isinstance(event, dict) else ""
        if not address:
            return {"ok": False, "error": "address is required"}
        valid, addr_type = validate_bitcoin_address(address)
        return {"ok": True, "valid": valid, "type": addr_type}
    except Exception as e:
        return {"ok": False, "error": str(e)}
