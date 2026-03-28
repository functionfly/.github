import ssl
import socket


def handler(event):
    """Check if a server supports HTTP/2 via ALPN."""
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
        ctx.set_alpn_protocols(["h2", "http/1.1"])

        with socket.create_connection((host, port), timeout=timeout) as sock:
            with ctx.wrap_socket(sock, server_hostname=host) as ssock:
                protocol = ssock.selected_alpn_protocol()

        supported = protocol == "h2"
        return {
            "ok": True,
            "supported": supported,
            "protocol": protocol or "http/1.1",
            "host": host,
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
