def handler(event):
    certificate = event.get("certificate") if isinstance(event, dict) else None

    if not certificate:
        return {"ok": False, "error": "certificate is required (PEM or Base64 DER)"}

    try:
        from cryptography import x509
        from cryptography.hazmat.primitives import serialization
        import base64

        cert_str = str(certificate).strip()
        if cert_str.startswith("-----"):
            cert = x509.load_pem_x509_certificate(cert_str.encode("utf-8"))
        else:
            der_bytes = base64.b64decode(cert_str)
            cert = x509.load_der_x509_certificate(der_bytes)

        subject = {attr.oid._name: attr.value for attr in cert.subject}
        issuer = {attr.oid._name: attr.value for attr in cert.issuer}

        pub_key = cert.public_key()
        key_info = {"algorithm": type(pub_key).__name__}

        return {
            "ok": True,
            "result": {
                "subject": subject,
                "issuer": issuer,
                "serial_number": str(cert.serial_number),
                "not_before": cert.not_valid_before_utc.isoformat(),
                "not_after": cert.not_valid_after_utc.isoformat(),
                "is_ca": False,
                "public_key": key_info,
                "version": str(cert.version.name),
                "fingerprint_sha256": cert.fingerprint(cert.signature_hash_algorithm).hex() if cert.signature_hash_algorithm else None,
            }
        }
    except ImportError:
        return {"ok": False, "error": "cryptography library is not installed. Install with: pip install cryptography"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
