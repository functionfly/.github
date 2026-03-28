def handler(event):
    """Generate an Apache HTTP server VirtualHost configuration."""
    try:
        server_name = event.get("server_name")
        if not server_name:
            return {"ok": False, "error": "server_name is required"}

        document_root = event.get("document_root", "/var/www/html")
        proxy_pass = event.get("proxy_pass")
        ssl = event.get("ssl", False)
        ssl_cert = event.get("ssl_cert", f"/etc/ssl/certs/{server_name}.crt")
        ssl_key = event.get("ssl_key", f"/etc/ssl/private/{server_name}.key")
        listen_port = event.get("listen_port", 443 if ssl else 80)

        lines = [f"<VirtualHost *:{listen_port}>"]
        lines.append(f"    ServerName {server_name}")

        if ssl:
            lines.extend([
                "    SSLEngine on",
                f"    SSLCertificateFile {ssl_cert}",
                f"    SSLCertificateKeyFile {ssl_key}",
            ])

        if proxy_pass:
            lines.extend([
                "    ProxyPreserveHost On",
                f"    ProxyPass / {proxy_pass}/",
                f"    ProxyPassReverse / {proxy_pass}/",
            ])
        else:
            lines.extend([
                f"    DocumentRoot {document_root}",
                f"    <Directory {document_root}>",
                "        Options Indexes FollowSymLinks",
                "        AllowOverride All",
                "        Require all granted",
                "    </Directory>",
            ])

        lines.extend([
            "    ErrorLog ${APACHE_LOG_DIR}/error.log",
            "    CustomLog ${APACHE_LOG_DIR}/access.log combined",
            "</VirtualHost>",
        ])

        config = "\n".join(lines) + "\n"
        return {"ok": True, "result": "Apache config generated", "config": config}
    except Exception as e:
        return {"ok": False, "error": str(e)}
