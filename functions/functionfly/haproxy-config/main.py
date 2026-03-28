def handler(event):
    """Generate an HAProxy configuration."""
    try:
        servers = event.get("servers")
        if not servers:
            return {"ok": False, "error": "servers is required"}

        frontend_name = event.get("frontend_name", "main")
        frontend_port = event.get("frontend_port", 80)
        backend_name = event.get("backend_name", "servers")
        balance = event.get("balance", "roundrobin")
        mode = event.get("mode", "http")

        lines = [
            "global",
            "    log /dev/log local0",
            "    log /dev/log local1 notice",
            "    maxconn 4096",
            "",
            "defaults",
            f"    mode {mode}",
            "    log global",
            "    option httplog",
            "    option dontlognull",
            "    timeout connect 5000ms",
            "    timeout client 50000ms",
            "    timeout server 50000ms",
            "",
            f"frontend {frontend_name}",
            f"    bind *:{frontend_port}",
            f"    default_backend {backend_name}",
            "",
            f"backend {backend_name}",
            f"    balance {balance}",
        ]

        if mode == "http":
            lines.append("    option httpchk GET /health")

        for srv in servers:
            name = srv.get("name", "server")
            host = srv.get("host", "127.0.0.1")
            port = srv.get("port", 8080)
            weight = srv.get("weight", 1)
            lines.append(f"    server {name} {host}:{port} weight {weight} check")

        config = "\n".join(lines) + "\n"
        return {"ok": True, "result": "HAProxy config generated", "config": config}
    except Exception as e:
        return {"ok": False, "error": str(e)}
