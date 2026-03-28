BECH32_ALPHABET = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

def bech32_polymod(values):
    """Compute Bech32 checksum"""
    GEN = [0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3]
    chk = 1
    for v in values:
        b = chk >> 25
        chk = (chk & 0x1ffffff) << 5 ^ v
        for i in range(5):
            chk ^= GEN[i] if ((b >> i) & 1) else 0
    return chk

def bech32_hrp_expand(hrp):
    """Expand HRP for checksum"""
    return [ord(x) >> 5 for x in hrp] + [0] + [ord(x) & 31 for x in hrp]

def bech32_verify_checksum(hrp, data):
    """Verify Bech32 checksum"""
    return bech32_polymod(bech32_hrp_expand(hrp) + data) == 1

def bech32_create_checksum(hrp, data):
    """Create Bech32 checksum"""
    values = bech32_hrp_expand(hrp) + data
    polymod = bech32_polymod(values + [0, 0, 0, 0, 0, 0]) ^ 1
    return [(polymod >> 5 * (5 - i)) & 31 for i in range(6)]

def encode_bech32(data, prefix="bc"):
    """Encode data to Bech32"""
    if not data:
        return ""
    # Convert data to 5-bit groups
    data_bytes = data.encode('utf-8')
    five_bit = []
    for byte in data_bytes:
        five_bit.append(byte >> 3)
        five_bit.append((byte & 7) << 2)
    # Remove trailing zeros
    while five_bit and five_bit[-1] == 0:
        five_bit.pop()
    # Create checksum
    checksum = bech32_create_checksum(prefix, five_bit)
    # Encode
    encoded = prefix + "1"
    for bit in five_bit:
        encoded += BECH32_ALPHABET[bit]
    for bit in checksum:
        encoded += BECH32_ALPHABET[bit]
    return encoded

def handler(event):
    try:
        data = event.get("data", "") if isinstance(event, dict) else ""
        prefix = event.get("prefix", "bc") if isinstance(event, dict) else "bc"
        if not data:
            return {"ok": False, "error": "data is required"}
        encoded = encode_bech32(data, prefix)
        return {"ok": True, "encoded": encoded}
    except Exception as e:
        return {"ok": False, "error": str(e)}
