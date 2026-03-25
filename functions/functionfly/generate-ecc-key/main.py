def handler(event):
    curve = event.get("curve", "P-256") if isinstance(event, dict) else "P-256"

    CURVES = ["P-256", "P-384", "P-521", "secp256k1"]
    if curve not in CURVES:
        return {"ok": False, "error": f"curve must be one of: {', '.join(CURVES)}"}

    try:
        from cryptography.hazmat.primitives.asymmetric import ec
        from cryptography.hazmat.primitives import serialization

        CURVE_MAP = {
            "P-256": ec.SECP256R1(),
            "P-384": ec.SECP384R1(),
            "P-521": ec.SECP521R1(),
            "secp256k1": ec.SECP256K1(),
        }

        private_key = ec.generate_private_key(CURVE_MAP[curve])
        public_key = private_key.public_key()

        priv_pem = private_key.private_bytes(
            encoding=serialization.Encoding.PEM,
            format=serialization.PrivateFormat.PKCS8,
            encryption_algorithm=serialization.NoEncryption()
        ).decode("utf-8")
        pub_pem = public_key.public_bytes(
            encoding=serialization.Encoding.PEM,
            format=serialization.PublicFormat.SubjectPublicKeyInfo
        ).decode("utf-8")

        return {"ok": True, "private_key": priv_pem, "public_key": pub_pem, "curve": curve}
    except ImportError:
        return {"ok": False, "error": "cryptography library is not installed. Install with: pip install cryptography"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
