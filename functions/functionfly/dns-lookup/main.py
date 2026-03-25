import socket


def handler(event):
    hostname = event.get("hostname") if isinstance(event, dict) else None
    record_type = event.get("record_type", "A")

    if not hostname:
        return {"ok": False, "error": "hostname is required"}
    if not isinstance(hostname, str):
        return {"ok": False, "error": "hostname must be a string"}

    hostname = hostname.strip()
    record_type = str(record_type).upper()

    try:
        if record_type in ("A", "AAAA", "ANY"):
            results = socket.getaddrinfo(hostname, None)
            addresses = list({r[4][0] for r in results})
            return {"ok": True, "hostname": hostname, "record_type": record_type, "addresses": addresses}
        elif record_type == "PTR":
            result = socket.gethostbyaddr(hostname)
            return {"ok": True, "hostname": hostname, "record_type": "PTR", "addresses": [result[0]]}
        else:
            # For MX, TXT, CNAME etc., fallback to A lookup
            ip = socket.gethostbyname(hostname)
            return {"ok": True, "hostname": hostname, "record_type": "A", "addresses": [ip]}
    except socket.gaierror as e:
        return {"ok": False, "error": f"DNS lookup failed: {str(e)}"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
