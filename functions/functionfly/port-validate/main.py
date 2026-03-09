def handler(event):
    port = event.get("port")

    if port is None:
        return {"ok": False, "error": "port is required"}

    try:
        # Convert to integer if it's a string
        if isinstance(port, str):
            port = int(port)
        elif not isinstance(port, int):
            return {"ok": False, "error": "port must be an integer or string representation of integer"}

        # Check valid range (0-65535)
        if port < 0 or port > 65535:
            return {
                "ok": False,
                "error": f"port {port} is out of valid range (0-65535)",
                "result": {
                    "port": port,
                    "is_valid": False,
                    "reason": "out_of_range"
                }
            }

        # Check for well-known ports
        well_known_ports = {
            20: "FTP Data",
            21: "FTP Control",
            22: "SSH",
            23: "Telnet",
            25: "SMTP",
            53: "DNS",
            80: "HTTP",
            110: "POP3",
            143: "IMAP",
            443: "HTTPS",
            993: "IMAPS",
            995: "POP3S"
        }

        result = {
            "port": port,
            "is_valid": True,
            "is_well_known": port in well_known_ports,
            "service": well_known_ports.get(port),
            "category": _get_port_category(port)
        }

        return {
            "ok": True,
            "result": result
        }

    except ValueError:
        return {
            "ok": False,
            "error": f"port '{port}' is not a valid integer",
            "result": {
                "port": port,
                "is_valid": False,
                "reason": "not_integer"
            }
        }
    except Exception as e:
        return {"ok": False, "error": f"failed to validate port: {str(e)}"}


def _get_port_category(port):
    """Categorize port number"""
    if port == 0:
        return "reserved"
    elif 1 <= port <= 1023:
        return "well_known"
    elif 1024 <= port <= 49151:
        return "registered"
    elif 49152 <= port <= 65535:
        return "dynamic"
    else:
        return "invalid"