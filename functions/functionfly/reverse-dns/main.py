import socket


def handler(event):
    ip = event.get("ip") if isinstance(event, dict) else None

    if not ip:
        return {"ok": False, "error": "ip is required"}
    if not isinstance(ip, str):
        return {"ok": False, "error": "ip must be a string"}

    ip = ip.strip()

    try:
        hostname, aliases, _ = socket.gethostbyaddr(ip)
        return {
            "ok": True,
            "ip": ip,
            "hostname": hostname,
            "aliases": aliases
        }
    except socket.herror as e:
        return {"ok": False, "error": f"Reverse DNS failed: {str(e)}"}
    except socket.gaierror as e:
        return {"ok": False, "error": f"DNS error: {str(e)}"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
