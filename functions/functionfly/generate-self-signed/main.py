import datetime


def handler(event):
    common_name = event.get("common_name", "localhost") if isinstance(event, dict) else "localhost"
    country = event.get("country", "US")
    org = event.get("org", "My Organization")
    valid_days = event.get("valid_days", 365)
    key_size = event.get("key_size", 2048)
    san_domains = event.get("san_domains", [])

    try:
        from cryptography import x509
        from cryptography.x509.oid import NameOID
        from cryptography.hazmat.primitives import hashes, serialization
        from cryptography.hazmat.primitives.asymmetric import rsa

        private_key = rsa.generate_private_key(public_exponent=65537, key_size=int(key_size))
        subject = issuer = x509.Name([
            x509.NameAttribute(NameOID.COUNTRY_NAME, str(country)[:2]),
            x509.NameAttribute(NameOID.ORGANIZATION_NAME, str(org)),
            x509.NameAttribute(NameOID.COMMON_NAME, str(common_name)),
        ])

        san_list = [x509.DNSName(str(common_name))]
        for d in (san_domains or []):
            san_list.append(x509.DNSName(str(d)))

        now = datetime.datetime.now(datetime.timezone.utc)
        cert = (
            x509.CertificateBuilder()
            .subject_name(subject)
            .issuer_name(issuer)
            .public_key(private_key.public_key())
            .serial_number(x509.random_serial_number())
            .not_valid_before(now)
            .not_valid_after(now + datetime.timedelta(days=int(valid_days)))
            .add_extension(x509.SubjectAlternativeName(san_list), critical=False)
            .add_extension(x509.BasicConstraints(ca=False, path_length=None), critical=True)
            .sign(private_key, hashes.SHA256())
        )

        cert_pem = cert.public_bytes(serialization.Encoding.PEM).decode("utf-8")
        key_pem = private_key.private_bytes(
            encoding=serialization.Encoding.PEM,
            format=serialization.PrivateFormat.TraditionalOpenSSL,
            encryption_algorithm=serialization.NoEncryption()
        ).decode("utf-8")

        return {"ok": True, "certificate": cert_pem, "private_key": key_pem, "common_name": common_name, "valid_days": int(valid_days)}
    except ImportError:
        return {"ok": False, "error": "cryptography library is not installed. Install with: pip install cryptography"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
