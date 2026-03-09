import re
import urllib.request
from html.parser import HTMLParser

class MetaParser(HTMLParser):
    def __init__(self):
        super().__init__()
        self.meta = {}
        self.title = ""
        self._in_title = False
    def handle_starttag(self, tag, attrs):
        if tag == "meta":
            d = dict(attrs)
            name = d.get("name") or d.get("property")
            content = d.get("content")
            if name and content:
                self.meta[name] = content
        elif tag == "title":
            self._in_title = True
    def handle_endtag(self, tag):
        if tag == "title":
            self._in_title = False
    def handle_data(self, data):
        if self._in_title:
            self.title = (self.title + data).strip()

def handler(event):
    url = event.get("url", "") if isinstance(event, dict) else ""
    if not url:
        return {"ok": False, "error": "url is required"}
    try:
        req = urllib.request.Request(url, headers={"User-Agent": "FunctionFly/1.0"})
        with urllib.request.urlopen(req, timeout=10) as r:
            html = r.read().decode("utf-8", errors="replace")
        p = MetaParser()
        p.feed(html)
        return {"ok": True, "meta": p.meta, "title": p.title}
    except Exception as e:
        return {"ok": False, "error": str(e)}
