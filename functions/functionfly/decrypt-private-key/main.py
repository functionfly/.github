def handler(event):
    private_key_pem = event.get("private_key") if isinstance(event, dict) else None
    passphrase = event.get("passphrase")

    if not private_key_pem:
        return {"ok": False, "error": "private_key is required"}
    if not passphrase:
        return {"ok": False, "error": "passphrase is required"}

    try:
        from cryptography.hazmat.primitives import serialization

        priv = serialization.load_pem_private_key(
            str(private_key_pem).encode("utf-8"),
            password=str(passphrase).encode("utf-8")
        )
        unencrypted_pem = priv.private_bytes(
            encoding=serialization.Encoding.PEM,
            format=serialization.PrivateFormat.PKCS8,
            encryption_algorithm=serialization.NoEncryption()
        ).decode("utf-8")

        return {"ok": True, "result": unencrypted_pem}
    except ImportError:
        return {"ok": False, "error": "cryptography library is not installed. Install with: pip install cryptography"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
