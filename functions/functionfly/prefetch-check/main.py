import urllib.request
import urllib.error
import ssl
import re


def handler(event):
    """Check for prefetch Link headers."""
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
                link_header = resp.headers.get("Link", "")
        except urllib.error.HTTPError as e:
            link_header = e.headers.get("Link", "")

        prefetches = []
        for part in link_header.split(","):
            part = part.strip()
            if "rel=prefetch" in part or 'rel="prefetch"' in part:
                match = re.match(r'<([^>]+)>', part)
                if match:
                    prefetches.append({"href": match.group(1)})

        return {"ok": True, "prefetches": prefetches, "count": len(prefetches), "url": url}
    except Exception as e:
        return {"ok": False, "error": str(e)}
