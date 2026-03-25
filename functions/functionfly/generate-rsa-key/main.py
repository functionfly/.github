def handler(event):
    key_size = event.get("key_size", 2048) if isinstance(event, dict) else 2048
    public_exponent = event.get("public_exponent", 65537)
    format_ = event.get("format", "pem")

    VALID_SIZES = [1024, 2048, 3072, 4096]
    try:
        key_size = int(key_size)
    except (TypeError, ValueError):
        return {"ok": False, "error": "key_size must be an integer"}
    if key_size not in VALID_SIZES:
        return {"ok": False, "error": f"key_size must be one of {VALID_SIZES}"}

    try:
        from cryptography.hazmat.primitives.asymmetric import rsa
        from cryptography.hazmat.primitives import serialization

        private_key = rsa.generate_private_key(
            public_exponent=int(public_exponent),
            key_size=key_size,
        )
        public_key = private_key.public_key()

        if format_ == "der":
            priv = private_key.private_bytes(
                encoding=serialization.Encoding.DER,
                format=serialization.PrivateFormat.PKCS8,
                encryption_algorithm=serialization.NoEncryption()
            )
            pub = public_key.public_bytes(
                encoding=serialization.Encoding.DER,
                format=serialization.PublicFormat.SubjectPublicKeyInfo
            )
            import base64
            return {"ok": True, "private_key": base64.b64encode(priv).decode(), "public_key": base64.b64encode(pub).decode(), "format": "der_base64", "key_size": key_size}
        else:
            priv_pem = private_key.private_bytes(
                encoding=serialization.Encoding.PEM,
                format=serialization.PrivateFormat.PKCS8,
                encryption_algorithm=serialization.NoEncryption()
            ).decode("utf-8")
            pub_pem = public_key.public_bytes(
                encoding=serialization.Encoding.PEM,
                format=serialization.PublicFormat.SubjectPublicKeyInfo
            ).decode("utf-8")
            return {"ok": True, "private_key": priv_pem, "public_key": pub_pem, "format": "pem", "key_size": key_size}
    except ImportError:
        return {"ok": False, "error": "cryptography library is not installed. Install with: pip install cryptography"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
