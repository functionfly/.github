import base64


def handler(event):
    key_material = event.get("key_material") if isinstance(event, dict) else None
    length = event.get("length", 32)
    info = event.get("info", "")
    salt = event.get("salt")
    algorithm = event.get("algorithm", "sha256")
    output = event.get("output", "hex")

    if not key_material:
        return {"ok": False, "error": "key_material is required"}

    try:
        from cryptography.hazmat.primitives import hashes
        from cryptography.hazmat.primitives.kdf.hkdf import HKDF

        HASH_MAP = {"sha256": hashes.SHA256(), "sha384": hashes.SHA384(), "sha512": hashes.SHA512()}
        if algorithm not in HASH_MAP:
            return {"ok": False, "error": f"algorithm must be one of: {', '.join(HASH_MAP.keys())}"}

        try:
            ikm = bytes.fromhex(str(key_material))
        except ValueError:
            ikm = base64.b64decode(str(key_material))

        salt_bytes = None
        if salt:
            try:
                salt_bytes = bytes.fromhex(str(salt))
            except ValueError:
                salt_bytes = base64.b64decode(str(salt))

        info_bytes = str(info).encode("utf-8") if info else b""

        hkdf = HKDF(algorithm=HASH_MAP[algorithm], length=int(length), salt=salt_bytes, info=info_bytes)
        dk = hkdf.derive(ikm)
        result = base64.b64encode(dk).decode("utf-8") if output == "base64" else dk.hex()
        return {"ok": True, "result": result, "algorithm": algorithm, "length": int(length)}
    except ImportError:
        return {"ok": False, "error": "cryptography library is not installed. Install with: pip install cryptography"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
