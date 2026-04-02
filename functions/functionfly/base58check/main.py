import hashlib

BASE58_ALPHABET = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"


def double_sha256(data):
    return hashlib.sha256(hashlib.sha256(data).digest()).digest()


def base58_encode(data_bytes):
    num = int.from_bytes(data_bytes, "big")
    result = []
    while num > 0:
        num, remainder = divmod(num, 58)
        result.append(BASE58_ALPHABET[remainder])
    for byte in data_bytes:
        if byte == 0:
            result.append("1")
        else:
            break
    return "".join(reversed(result))


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    data = event.get("data")
    if not data:
        return {"ok": False, "error": "data is required"}
    try:
        version = int(event.get("version", 0))
        payload = bytes.fromhex(data.replace(" ", ""))
        version_byte = bytes([version])
        payload_with_version = version_byte + payload
        checksum = double_sha256(payload_with_version)[:4]
        full = payload_with_version + checksum
        encoded = base58_encode(full)
        return {"ok": True, "encoded": encoded}
    except ValueError:
        return {"ok": False, "error": "data must be valid hex string"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
