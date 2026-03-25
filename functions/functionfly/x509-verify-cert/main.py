from datetime import datetime, timezone


def handler(event):
    certificate = event.get("certificate") if isinstance(event, dict) else None
    ca_certificate = event.get("ca_certificate")
    check_expiry = event.get("check_expiry", True)

    if not certificate:
        return {"ok": False, "error": "certificate is required"}

    try:
        from cryptography import x509
        from cryptography.hazmat.primitives.asymmetric import padding
        from cryptography.hazmat.primitives import hashes
        import base64

        cert_str = str(certificate).strip()
        if cert_str.startswith("-----"):
            cert = x509.load_pem_x509_certificate(cert_str.encode("utf-8"))
        else:
            cert = x509.load_der_x509_certificate(base64.b64decode(cert_str))

        now = datetime.now(timezone.utc)
        is_expired = now > cert.not_valid_after_utc
        is_not_yet_valid = now < cert.not_valid_before_utc

        checks = {
            "expired": is_expired,
            "not_yet_valid": is_not_yet_valid,
            "valid_period": not is_expired and not is_not_yet_valid,
        }

        signature_valid = None
        if ca_certificate:
            try:
                ca_str = str(ca_certificate).strip()
                if ca_str.startswith("-----"):
                    ca_cert = x509.load_pem_x509_certificate(ca_str.encode("utf-8"))
                else:
                    ca_cert = x509.load_der_x509_certificate(base64.b64decode(ca_str))
                ca_pub = ca_cert.public_key()
                ca_pub.verify(cert.signature, cert.tbs_certificate_bytes, padding.PKCS1v15(), cert.signature_hash_algorithm)
                signature_valid = True
            except Exception:
                signature_valid = False
            checks["signature_valid"] = signature_valid

        is_valid = checks["valid_period"] and (signature_valid is not False)

        return {
            "ok": True,
            "result": is_valid,
            "checks": checks,
            "not_before": cert.not_valid_before_utc.isoformat(),
            "not_after": cert.not_valid_after_utc.isoformat(),
        }
    except ImportError:
        return {"ok": False, "error": "cryptography library is not installed. Install with: pip install cryptography"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
