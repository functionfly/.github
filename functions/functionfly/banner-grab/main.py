import socket


SERVICE_PATTERNS = {
    "SSH": "SSH-",
    "FTP": "220 ",
    "SMTP": "220 ",
    "HTTP": "HTTP/",
    "POP3": "+OK",
    "IMAP": "* OK",
    "MySQL": "\x4a\x00\x00\x00",
    "Redis": "+PONG",
}


def detect_service(banner, port):
    for service, pattern in SERVICE_PATTERNS.items():
        if banner.startswith(pattern):
            return service
    # Port-based fallback
    port_services = {21: "FTP", 22: "SSH", 25: "SMTP", 80: "HTTP", 110: "POP3", 143: "IMAP", 443: "HTTPS", 3306: "MySQL", 5432: "PostgreSQL", 6379: "Redis"}
    return port_services.get(port, "Unknown")


def handler(event):
    """Grab the service banner from a TCP port."""
    try:
        host = event.get("host")
        port = event.get("port")
        if not host or port is None:
            return {"ok": False, "error": "host and port are required"}

        host = host.replace("https://", "").replace("http://", "").split("/")[0]
        port = int(port)
        timeout = float(event.get("timeout", 5))

        with socket.create_connection((host, port), timeout=timeout) as sock:
            sock.settimeout(timeout)
            try:
                # Send HTTP request for web ports
                if port in (80, 443, 8080, 8443):
                    sock.send(b"HEAD / HTTP/1.0\r\n\r\n")
                banner_bytes = sock.recv(1024)
                banner = banner_bytes.decode("utf-8", errors="replace").strip()
            except socket.timeout:
                banner = ""

        service = detect_service(banner, port)
        return {"ok": True, "banner": banner, "service": service, "host": host, "port": port}
    except Exception as e:
        return {"ok": False, "error": str(e)}
