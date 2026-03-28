def handler(event):
    """Generate a Caddyfile configuration."""
    try:
        domain = event.get("domain")
        if not domain:
            return {"ok": False, "error": "domain is required"}

        proxy_pass = event.get("proxy_pass")
        root = event.get("root")
        tls_email = event.get("tls_email")
        headers = event.get("headers", {})
        routes = event.get("routes", [])

        lines = []

        if tls_email:
            lines.extend([f"{{", f"    email {tls_email}", "}", ""])

        lines.append(f"{domain} {{")

        if proxy_pass:
            lines.append(f"    reverse_proxy {proxy_pass}")
        elif root:
            lines.append(f"    root * {root}")
            lines.append("    file_server")

        for header_name, header_value in headers.items():
            lines.append(f"    header {header_name} {header_value}")

        for route in routes:
            path = route.get("path", "/")
            if route.get("proxy_pass"):
                lines.append(f"    handle {path} {{")
                lines.append(f"        reverse_proxy {route['proxy_pass']}")
                lines.append("    }")
            elif route.get("root"):
                lines.append(f"    handle {path} {{")
                lines.append(f"        root * {route['root']}")
                lines.append("        file_server")
                lines.append("    }")

        lines.append("}")

        caddyfile = "\n".join(lines) + "\n"
        return {"ok": True, "result": "Caddyfile generated", "caddyfile": caddyfile}
    except Exception as e:
        return {"ok": False, "error": str(e)}
