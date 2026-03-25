import re


def handler(event):
    xml = event.get("xml") if isinstance(event, dict) else None
    max_urls = int(event.get("max_urls", 100))
    if not xml:
        return {"ok": False, "error": "xml is required"}
    try:
        text = str(xml)
        is_index = bool(re.search(r'<sitemapindex', text, re.I))
        if is_index:
            sitemaps = []
            for m in re.finditer(r'<sitemap>(.*?)</sitemap>', text, re.S | re.I):
                s = m.group(1)
                loc_m = re.search(r'<loc[^>]*>(.*?)</loc>', s, re.I)
                lastmod_m = re.search(r'<lastmod[^>]*>(.*?)</lastmod>', s, re.I)
                if loc_m:
                    sitemaps.append({"loc": loc_m.group(1).strip(), "lastmod": lastmod_m.group(1).strip() if lastmod_m else None})
            return {"ok": True, "result": sitemaps, "type": "sitemap_index", "sitemaps": sitemaps, "count": len(sitemaps)}
        urls = []
        for m in re.finditer(r'<url>(.*?)</url>', text, re.S | re.I):
            u = m.group(1)
            loc_m = re.search(r'<loc[^>]*>(.*?)</loc>', u, re.I)
            if loc_m:
                urls.append({
                    "loc": loc_m.group(1).strip(),
                    "lastmod": (re.search(r'<lastmod[^>]*>(.*?)</lastmod>', u, re.I) or type('', (), {'group': lambda self, x: None})()).group(1),
                    "changefreq": (re.search(r'<changefreq[^>]*>(.*?)</changefreq>', u, re.I) or type('', (), {'group': lambda self, x: None})()).group(1),
                    "priority": (re.search(r'<priority[^>]*>(.*?)</priority>', u, re.I) or type('', (), {'group': lambda self, x: None})()).group(1),
                })
            if len(urls) >= max_urls:
                break
        return {"ok": True, "result": urls, "type": "urlset", "urls": urls, "count": len(urls)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
