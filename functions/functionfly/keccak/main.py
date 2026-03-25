import hashlib


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    variant = event.get("variant", "sha3_256")

    if data is None:
        return {"ok": False, "error": "data is required"}

    VARIANTS = {
        "sha3_224": "sha3_224",
        "sha3_256": "sha3_256",
        "keccak_256": "sha3_256",
        "sha3_384": "sha3_384",
        "sha3_512": "sha3_512",
        "shake_128": "shake_128",
        "shake_256": "shake_256",
    }

    try:
        if isinstance(data, (dict, list)):
            import json
            raw = json.dumps(data).encode("utf-8")
        else:
            raw = str(data).encode("utf-8")

        algo = VARIANTS.get(variant, "sha3_256")
        if algo in ("shake_128", "shake_256"):
            length = event.get("length", 32)
            h = hashlib.new(algo, raw)
            digest = h.hexdigest(int(length))
        else:
            h = hashlib.new(algo, raw)
            digest = h.hexdigest()

        return {"ok": True, "result": digest, "algorithm": algo, "variant": variant, "digest_bits": len(digest) * 4}
    except ValueError as e:
        return {"ok": False, "error": f"Unsupported variant '{variant}': {e}"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
