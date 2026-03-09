import re
import urllib.request
from html.parser import HTMLParser

class TextExtractor(HTMLParser):
    def __init__(self):
        super().__init__()
        self.skip = False
        self._in_title = False
        self.text = []
        self.title = ""
    def handle_starttag(self, tag, attrs):
        if tag in ("script", "style", "nav", "footer", "head"):
            self.skip = True
        elif tag == "title":
            self._in_title = True
    def handle_endtag(self, tag):
        if tag in ("script", "style", "nav", "footer", "head"):
            self.skip = False
        elif tag == "title":
            self._in_title = False
    def handle_data(self, data):
        if self._in_title:
            self.title = (self.title + data).strip()
        elif not self.skip and data:
            self.text.append(data.strip())

def handler(event):
    url = event.get("url", "") if isinstance(event, dict) else ""
    if not url:
        return {"ok": False, "error": "url is required"}
    try:
        req = urllib.request.Request(url, headers={"User-Agent": "FunctionFly/1.0"})
        with urllib.request.urlopen(req, timeout=12) as r:
            html = r.read().decode("utf-8", errors="replace")
        p = TextExtractor()
        p.feed(html)
        title = p.title
        text = " ".join(t for t in p.text if t)
        text = re.sub(r"\s+", " ", text).strip()
        return {"ok": True, "text": text[:50000], "title": title}
    except Exception as e:
        return {"ok": False, "error": str(e)}
