import re


def handler(event):
    xml = event.get("xml") if isinstance(event, dict) else None
    max_items = int(event.get("max_items", 20))
    if not xml:
        return {"ok": False, "error": "xml is required"}
    try:
        text = str(xml)
        def tag(name, s):
            m = re.search(rf'<(?:[a-z]+:)?{name}[^>]*>(.*?)</(?:[a-z]+:)?{name}>', s, re.S | re.I)
            return m.group(1).strip() if m else None
        def get_attr(name, attr, s):
            m = re.search(rf'<(?:[a-z]+:)?{name}[^>]+{attr}=["\']([^"\']+)["\']', s, re.I)
            return m.group(1) if m else None
        title = tag("title", text)
        subtitle = tag("subtitle", text)
        updated = tag("updated", text)
        entries = []
        for entry_m in re.finditer(r'<entry[^>]*>(.*?)</entry>', text, re.S | re.I):
            e = entry_m.group(1)
            link = get_attr("link", "href", e)
            entries.append({
                "title": tag("title", e),
                "link": link,
                "summary": tag("summary", e),
                "content": tag("content", e),
                "published": tag("published", e),
                "updated": tag("updated", e),
                "id": tag("id", e),
                "author": tag("name", e),
            })
            if len(entries) >= max_items:
                break
        return {"ok": True, "result": entries, "feed": {"title": title, "subtitle": subtitle, "updated": updated}, "entries": entries, "count": len(entries)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
