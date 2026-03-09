def handler(event):
    if isinstance(event, dict):
        bits = event.get("bits", 2048)
        fmt = (event.get("format") or "pem").lower()
    else:
        bits, fmt = 2048, "pem"

    try:
        bits = max(1024, min(4096, int(bits)))
    except (TypeError, ValueError):
        bits = 2048

    try:
        from cryptography.hazmat.primitives.asymmetric import rsa
        from cryptography.hazmat.primitives import serialization
    except ImportError:
        return {"ok": False, "error": "cryptography required: pip install cryptography"}

    try:
        key = rsa.generate_private_key(public_exponent=65537, key_size=bits)
        priv_pem = key.private_bytes(
            serialization.Encoding.PEM,
            serialization.PrivateFormat.PKCS8,
            serialization.NoEncryption(),
        )
        pub = key.public_key()
        pub_pem = pub.public_bytes(
            serialization.Encoding.PEM,
            serialization.PublicFormat.SubjectPublicKeyInfo,
        )
        if fmt == "der":
            priv_out = key.private_bytes(serialization.Encoding.DER, serialization.PrivateFormat.PKCS8, serialization.NoEncryption())
            pub_out = pub.public_bytes(serialization.Encoding.DER, serialization.PublicFormat.SubjectPublicKeyInfo)
            import base64
            return {"ok": True, "private_key": base64.b64encode(priv_out).decode("ascii"), "public_key": base64.b64encode(pub_out).decode("ascii")}
        return {"ok": True, "private_key": priv_pem.decode("utf-8"), "public_key": pub_pem.decode("utf-8")}
    except Exception as e:
        return {"ok": False, "error": str(e)}
