import ssl
import socket
import time


def handler(event):
    """Test SSL/TLS handshake."""
    try:
        host = event.get("host")
        if not host:
            return {"ok": False, "error": "host is required"}

        host = host.replace("https://", "").replace("http://", "").split("/")[0]
        port = int(event.get("port", 443))
        timeout = int(event.get("timeout", 10))

        ctx = ssl.create_default_context()
        ctx.check_hostname = False
        ctx.verify_mode = ssl.CERT_NONE

        start = time.time()
        try:
            with socket.create_connection((host, port), timeout=timeout) as sock:
                with ctx.wrap_socket(sock, server_hostname=host) as ssock:
                    handshake_ms = round((time.time() - start) * 1000, 2)
                    cipher = ssock.cipher()
                    protocol = ssock.version()

            return {
                "ok": True,
                "success": True,
                "protocol": protocol,
                "cipher": cipher[0] if cipher else None,
                "cipher_bits": cipher[2] if cipher else None,
                "handshake_ms": handshake_ms,
                "host": host,
                "port": port,
            }
        except ssl.SSLError as e:
            return {"ok": True, "success": False, "error": str(e), "host": host}
    except Exception as e:
        return {"ok": False, "error": str(e)}
