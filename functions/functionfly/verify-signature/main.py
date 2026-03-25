"""Generic signature verifier that auto-detects RSA/ECDSA key type."""
import base64


def handler(event):
    message = event.get("message") if isinstance(event, dict) else None
    signature = event.get("signature")
    public_key_pem = event.get("public_key")
    algorithm = event.get("algorithm", "sha256")
    scheme = event.get("scheme", "auto")  # 'auto', 'rsa_pkcs1', 'rsa_pss', 'ecdsa'

    if message is None:
        return {"ok": False, "error": "message is required"}
    if not signature:
        return {"ok": False, "error": "signature is required"}
    if not public_key_pem:
        return {"ok": False, "error": "public_key is required"}

    try:
        from cryptography.hazmat.primitives import hashes, serialization
        from cryptography.hazmat.primitives.asymmetric import rsa, ec, padding
        from cryptography.exceptions import InvalidSignature

        HASH_MAP = {
            "sha1": hashes.SHA1(),
            "sha256": hashes.SHA256(),
            "sha384": hashes.SHA384(),
            "sha512": hashes.SHA512(),
        }
        if algorithm not in HASH_MAP:
            return {"ok": False, "error": f"algorithm must be one of: {', '.join(HASH_MAP.keys())}"}

        pub = serialization.load_pem_public_key(str(public_key_pem).encode("utf-8"))
        sig_bytes = base64.b64decode(str(signature))

        if isinstance(message, (dict, list)):
            import json
            msg_bytes = json.dumps(message).encode("utf-8")
        else:
            msg_bytes = str(message).encode("utf-8")

        key_type = type(pub).__name__
        try:
            if isinstance(pub, ec.EllipticCurvePublicKey):
                pub.verify(sig_bytes, msg_bytes, ec.ECDSA(HASH_MAP[algorithm]))
                detected_scheme = "ecdsa"
            elif scheme == "rsa_pss":
                pub.verify(sig_bytes, msg_bytes, padding.PSS(mgf=padding.MGF1(HASH_MAP[algorithm]), salt_length=padding.PSS.MAX_LENGTH), HASH_MAP[algorithm])
                detected_scheme = "rsa_pss"
            else:
                pub.verify(sig_bytes, msg_bytes, padding.PKCS1v15(), HASH_MAP[algorithm])
                detected_scheme = "rsa_pkcs1"
            return {"ok": True, "result": True, "scheme": detected_scheme, "key_type": key_type}
        except InvalidSignature:
            return {"ok": True, "result": False, "key_type": key_type}
    except ImportError:
        return {"ok": False, "error": "cryptography library is not installed. Install with: pip install cryptography"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
