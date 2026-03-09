import re
import urllib.request
from html.parser import HTMLParser
from urllib.parse import urljoin, urlparse

class FaviconParser(HTMLParser):
    def __init__(self, base_url):
        super().__init__()
        self.base = base_url
        self.favicon = None
    def handle_starttag(self, tag, attrs):
        if tag != "link":
            return
        d = dict(attrs)
        rel = (d.get("rel") or "").lower()
        if "icon" in rel:
            href = d.get("href")
            if href:
                self.favicon = urljoin(self.base, href)

def handler(event):
    if isinstance(event, dict):
        url = event.get("url", "") or event.get("domain", "")
    else:
        url = ""
    if not url:
        return {"ok": False, "error": "url or domain is required"}
    if "://" not in url:
        url = "https://" + url
    parsed = urlparse(url)
    base = f"{parsed.scheme}://{parsed.netloc}"
    try:
        req = urllib.request.Request(url, headers={"User-Agent": "FunctionFly/1.0"})
        with urllib.request.urlopen(req, timeout=8) as r:
            html = r.read().decode("utf-8", errors="replace")
        p = FaviconParser(url)
        p.feed(html)
        if p.favicon:
            return {"ok": True, "favicon_url": p.favicon}
        return {"ok": True, "favicon_url": base + "/favicon.ico"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
