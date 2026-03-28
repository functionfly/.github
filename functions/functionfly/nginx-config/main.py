def handler(event):
    """Generate an Nginx server configuration."""
    try:
        server_name = event.get("server_name")
        if not server_name:
            return {"ok": False, "error": "server_name is required"}

        listen_port = event.get("listen_port", 80)
        root = event.get("root")
        proxy_pass = event.get("proxy_pass")
        ssl = event.get("ssl", False)
        ssl_cert = event.get("ssl_cert", f"/etc/ssl/certs/{server_name}.crt")
        ssl_key = event.get("ssl_key", f"/etc/ssl/private/{server_name}.key")
        locations = event.get("locations", [])
        gzip = event.get("gzip", True)

        lines = []

        if gzip:
            lines.extend([
                "gzip on;",
                "gzip_types text/plain text/css application/json application/javascript text/xml application/xml;",
                "",
            ])

        lines.append("server {")

        if ssl:
            lines.append(f"    listen 443 ssl;")
            lines.append(f"    ssl_certificate {ssl_cert};")
            lines.append(f"    ssl_certificate_key {ssl_key};")
            lines.append(f"    ssl_protocols TLSv1.2 TLSv1.3;")
        else:
            lines.append(f"    listen {listen_port};")

        lines.append(f"    server_name {server_name};")

        if root:
            lines.append(f"    root {root};")
            lines.append("    index index.html index.htm;")

        if proxy_pass and not locations:
            lines.extend([
                "",
                "    location / {",
                f"        proxy_pass {proxy_pass};",
                "        proxy_set_header Host $host;",
                "        proxy_set_header X-Real-IP $remote_addr;",
                "        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;",
                "        proxy_set_header X-Forwarded-Proto $scheme;",
                "    }",
            ])
        elif root and not locations:
            lines.extend([
                "",
                "    location / {",
                "        try_files $uri $uri/ =404;",
                "    }",
            ])

        for loc in locations:
            path = loc.get("path", "/")
            lines.append(f"\n    location {path} {{")
            if loc.get("proxy_pass"):
                lines.append(f"        proxy_pass {loc['proxy_pass']};")
                lines.append("        proxy_set_header Host $host;")
            if loc.get("root"):
                lines.append(f"        root {loc['root']};")
            if loc.get("try_files"):
                lines.append(f"        try_files {loc['try_files']};")
            lines.append("    }")

        lines.append("}")

        if ssl:
            lines.extend([
                "",
                "# HTTP to HTTPS redirect",
                "server {",
                f"    listen {listen_port};",
                f"    server_name {server_name};",
                "    return 301 https://$host$request_uri;",
                "}",
            ])

        config = "\n".join(lines) + "\n"
        return {"ok": True, "result": "Nginx config generated", "config": config}
    except Exception as e:
        return {"ok": False, "error": str(e)}
