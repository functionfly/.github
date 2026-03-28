import socket


def check_port(host, port, timeout):
    try:
        with socket.create_connection((host, port), timeout=timeout):
            return True
    except (socket.timeout, ConnectionRefusedError, OSError):
        return False


def handler(event):
    """Check if specific ports are open on a host."""
    try:
        host = event.get("host")
        ports = event.get("ports")
        if not host or not ports:
            return {"ok": False, "error": "host and ports are required"}

        host = host.replace("https://", "").replace("http://", "").split("/")[0]
        timeout = float(event.get("timeout", 3))
        ports = [int(p) for p in ports[:50]]  # Limit to 50 ports

        results = []
        open_ports = []

        for port in ports:
            is_open = check_port(host, port, timeout)
            results.append({"port": port, "open": is_open})
            if is_open:
                open_ports.append(port)

        return {"ok": True, "host": host, "results": results, "open_ports": open_ports}
    except Exception as e:
        return {"ok": False, "error": str(e)}
