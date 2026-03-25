import re


def handler(event):
    xml = event.get("xml") if isinstance(event, dict) else None
    max_items = int(event.get("max_items", 20))
    if not xml:
        return {"ok": False, "error": "xml is required"}
    try:
        text = str(xml)
        def tag(name, s):
            m = re.search(rf'<{name}[^>]*>(.*?)</{name}>', s, re.S | re.I)
            return m.group(1).strip() if m else None
        def strip_cdata(s):
            if not s: return s
            m = re.match(r'<!\[CDATA\[(.*?)\]\]>', s, re.S)
            return m.group(1) if m else s
        title = strip_cdata(tag("title", text))
        link = tag("link", text)
        description = strip_cdata(tag("description", text))
        items = []
        for item_m in re.finditer(r'<item>(.*?)</item>', text, re.S | re.I):
            item_xml = item_m.group(1)
            items.append({
                "title": strip_cdata(tag("title", item_xml)),
                "link": tag("link", item_xml),
                "description": strip_cdata(tag("description", item_xml)),
                "pubDate": tag("pubDate", item_xml),
                "guid": tag("guid", item_xml),
                "author": tag("author", item_xml) or tag("dc:creator", item_xml),
            })
            if len(items) >= max_items:
                break
        return {"ok": True, "result": items, "feed": {"title": title, "link": link, "description": description}, "items": items, "count": len(items)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
