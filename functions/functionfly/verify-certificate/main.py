import ssl
import socket


def handler(event):
    if isinstance(event, dict):
        hostname = (event.get("hostname") or event.get("host") or "").strip()
        port = event.get("port", 443)
    else:
        hostname, port = "", 443

    if not hostname:
        return {"ok": False, "error": "Input 'hostname' is required"}

    try:
        port = int(port)
    except (TypeError, ValueError):
        port = 443

    try:
        ctx = ssl.create_default_context()
        with socket.create_connection((hostname, port), timeout=10) as sock:
            with ctx.wrap_socket(sock, server_hostname=hostname) as ssock:
                cert = ssock.getpeercert()
                import datetime
                not_after = cert.get("notAfter")
                issuer = dict(x[0] for x in cert.get("issuer", []))
                issuer_str = issuer.get("organizationName", "") or issuer.get("commonName", "")
                if not_after:
                    import datetime as dt
                    exp = dt.datetime.strptime(not_after, "%b %d %H:%M:%S %Y %Z")
                    valid = exp.timestamp() > __import__("time").time()
                else:
                    valid = True
                    exp = None
                return {"ok": True, "valid": valid, "expires_at": not_after, "issuer": issuer_str}
    except ssl.SSLCertVerificationError as e:
        return {"ok": True, "valid": False, "error": str(e)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
