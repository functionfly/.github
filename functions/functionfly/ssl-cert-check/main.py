import ssl
import socket
from datetime import datetime, timezone


def handler(event):
    """Check SSL/TLS certificate details for a domain."""
    try:
        host = event.get("host")
        if not host:
            return {"ok": False, "error": "host is required"}

        host = host.replace("https://", "").replace("http://", "").split("/")[0]
        port = int(event.get("port", 443))
        timeout = int(event.get("timeout", 10))

        ctx = ssl.create_default_context()
        try:
            with socket.create_connection((host, port), timeout=timeout) as sock:
                with ctx.wrap_socket(sock, server_hostname=host) as ssock:
                    cert = ssock.getpeercert()
        except ssl.SSLCertVerificationError as e:
            # Try without verification to get cert info
            ctx2 = ssl.create_default_context()
            ctx2.check_hostname = False
            ctx2.verify_mode = ssl.CERT_NONE
            with socket.create_connection((host, port), timeout=timeout) as sock:
                with ctx2.wrap_socket(sock, server_hostname=host) as ssock:
                    cert = ssock.getpeercert()
            cert["_verification_error"] = str(e)

        # Parse expiry
        not_after = cert.get("notAfter")
        expires_at = None
        days_until_expiry = None
        if not_after:
            exp = datetime.strptime(not_after, "%b %d %H:%M:%S %Y %Z").replace(tzinfo=timezone.utc)
            expires_at = exp.isoformat()
            days_until_expiry = (exp - datetime.now(timezone.utc)).days

        # Parse subject and issuer
        subject = dict(x[0] for x in cert.get("subject", []))
        issuer = dict(x[0] for x in cert.get("issuer", []))

        valid = days_until_expiry is not None and days_until_expiry > 0 and "_verification_error" not in cert

        return {
            "ok": True,
            "valid": valid,
            "expires_at": expires_at,
            "days_until_expiry": days_until_expiry,
            "subject": subject,
            "issuer": issuer.get("organizationName", str(issuer)),
            "san": cert.get("subjectAltName"),
            "host": host,
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
