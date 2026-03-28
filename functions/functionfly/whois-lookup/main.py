import socket
import re


WHOIS_SERVERS = {
    "com": "whois.verisign-grs.com",
    "net": "whois.verisign-grs.com",
    "org": "whois.pir.org",
    "io": "whois.nic.io",
    "co": "whois.nic.co",
    "uk": "whois.nic.uk",
    "de": "whois.denic.de",
    "fr": "whois.nic.fr",
    "default": "whois.iana.org",
}


def handler(event):
    """Perform a WHOIS lookup for a domain."""
    try:
        domain = event.get("domain")
        if not domain:
            return {"ok": False, "error": "domain is required"}

        domain = domain.replace("https://", "").replace("http://", "").split("/")[0].lower()
        timeout = int(event.get("timeout", 10))

        tld = domain.split(".")[-1]
        server = WHOIS_SERVERS.get(tld, WHOIS_SERVERS["default"])

        with socket.create_connection((server, 43), timeout=timeout) as sock:
            sock.send(f"{domain}\r\n".encode())
            response = b""
            while True:
                data = sock.recv(4096)
                if not data:
                    break
                response += data

        raw = response.decode("utf-8", errors="replace")

        # Extract key fields
        registrar = None
        expiry = None
        for line in raw.splitlines():
            line_lower = line.lower()
            if "registrar:" in line_lower and not registrar:
                registrar = line.split(":", 1)[-1].strip()
            elif any(x in line_lower for x in ["expiry date:", "expiration date:", "registry expiry date:"]) and not expiry:
                expiry = line.split(":", 1)[-1].strip()

        return {
            "ok": True,
            "domain": domain,
            "result": raw[:2000],  # Truncate to 2000 chars
            "registrar": registrar,
            "expiry": expiry,
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
