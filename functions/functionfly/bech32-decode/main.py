BECH32_ALPHABET = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"


def bech32_polymod(values):
    GEN = [0x3B6A57B2, 0x26508E6D, 0x1EA119FA, 0x3D4233DD, 0x2A1462B3]
    chk = 1
    for v in values:
        b = chk >> 25
        chk = (chk & 0x1FFFFFF) << 5 ^ v
        for i in range(5):
            chk ^= GEN[i] if ((b >> i) & 1) else 0
    return chk


def bech32_hrp_expand(hrp):
    return [ord(x) >> 5 for x in hrp] + [0] + [ord(x) & 31 for x in hrp]


def bech32_verify_checksum(hrp, data):
    return bech32_polymod(bech32_hrp_expand(hrp) + data) == 1


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    data = event.get("data")
    if not data or not isinstance(data, str):
        return {"ok": False, "error": "data (Bech32 string) is required"}
    try:
        encoded = data.lower()
        if "1" not in encoded:
            return {"ok": False, "error": "invalid Bech32: no separator"}
        pos = encoded.rfind("1")
        prefix = encoded[:pos]
        data_part = encoded[pos + 1 :]
        if len(data_part) < 6:
            return {"ok": False, "error": "invalid Bech32: data too short"}
        values = [BECH32_ALPHABET.index(c) for c in data_part]
        if not bech32_verify_checksum(prefix, values):
            return {"ok": False, "error": "invalid checksum"}
        # Strip checksum (last 6 chars)
        data_values = values[:-6]
        # Convert 5-bit groups back to bytes
        bits = 0
        acc = 0
        result_bytes = []
        for v in data_values:
            acc = (acc << 5) | v
            bits += 5
            while bits >= 8:
                bits -= 8
                result_bytes.append((acc >> bits) & 0xFF)
        decoded = bytes(result_bytes).decode("utf-8")
        return {"ok": True, "decoded": decoded, "prefix": prefix}
    except (ValueError, IndexError):
        return {"ok": False, "error": "invalid Bech32 characters"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
