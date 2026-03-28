import urllib.request
import urllib.error
import ssl
import re


def parse_link_header(link_header):
    """Parse Link header into resource hints."""
    hints = []
    if not link_header:
        return hints
    for part in link_header.split(","):
        part = part.strip()
        match = re.match(r'<([^>]+)>(.*)', part)
        if not match:
            continue
        href = match.group(1)
        attrs_str = match.group(2)
        attrs = {}
        for attr in re.findall(r';\s*([a-z-]+)(?:="([^"]*)")?', attrs_str):
            attrs[attr[0]] = attr[1] if attr[1] else True
        rel = attrs.get("rel", "")
        if rel in ("preload", "prefetch", "preconnect", "dns-prefetch", "modulepreload"):
            hint = {"rel": rel, "href": href}
            if attrs.get("as"):
                hint["as"] = attrs["as"]
            hints.append(hint)
    return hints


def handler(event):
    """Check for resource hints in HTTP headers."""
    try:
        url = event.get("url")
        if not url:
            return {"ok": False, "error": "url is required"}

        timeout = int(event.get("timeout", 10))

        ctx = ssl.create_default_context()
        ctx.check_hostname = False
        ctx.verify_mode = ssl.CERT_NONE

        try:
            req = urllib.request.Request(url, method="HEAD")
            with urllib.request.urlopen(req, timeout=timeout, context=ctx) as resp:
                headers = resp.headers
                link_header = headers.get("Link")
        except urllib.error.HTTPError as e:
            link_header = e.headers.get("Link")

        hints = parse_link_header(link_header)

        return {
            "ok": True,
            "hints": hints,
            "count": len(hints),
            "link_header": link_header,
            "url": url,
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
