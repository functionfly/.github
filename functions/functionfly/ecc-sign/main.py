import base64


def handler(event):
    message = event.get("message") if isinstance(event, dict) else None
    private_key_pem = event.get("private_key")
    algorithm = event.get("algorithm", "sha256")

    if message is None:
        return {"ok": False, "error": "message is required"}
    if not private_key_pem:
        return {"ok": False, "error": "private_key is required"}

    try:
        from cryptography.hazmat.primitives import hashes, serialization
        from cryptography.hazmat.primitives.asymmetric import ec

        HASH_MAP = {"sha256": hashes.SHA256(), "sha384": hashes.SHA384(), "sha512": hashes.SHA512()}
        if algorithm not in HASH_MAP:
            return {"ok": False, "error": f"algorithm must be one of: {', '.join(HASH_MAP.keys())}"}

        priv = serialization.load_pem_private_key(str(private_key_pem).encode("utf-8"), password=None)

        if isinstance(message, (dict, list)):
            import json
            msg_bytes = json.dumps(message).encode("utf-8")
        else:
            msg_bytes = str(message).encode("utf-8")

        signature = priv.sign(msg_bytes, ec.ECDSA(HASH_MAP[algorithm]))
        return {"ok": True, "result": base64.b64encode(signature).decode("utf-8"), "algorithm": algorithm}
    except ImportError:
        return {"ok": False, "error": "cryptography library is not installed. Install with: pip install cryptography"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
