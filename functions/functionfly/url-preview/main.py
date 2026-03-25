import re


def handler(event):
    url = event.get("url") if isinstance(event, dict) else None
    html = event.get("html")
    if not url:
        return {"ok": False, "error": "url is required"}
    try:
        from urllib.parse import urlparse
        parsed = urlparse(str(url))
        domain = parsed.netloc.replace("www.", "")
        path_parts = [p for p in parsed.path.split("/") if p]
        slug = path_parts[-1] if path_parts else ""
        readable_slug = re.sub(r'[-_]', ' ', slug).title() if slug else ""
        preview = {
            "url": str(url),
            "domain": domain,
            "scheme": parsed.scheme,
            "path": parsed.path,
            "slug": slug,
            "readable_slug": readable_slug,
            "is_https": parsed.scheme == "https",
            "has_path": bool(path_parts),
            "has_query": bool(parsed.query)
        }
        if html:
            def extract(prop_attr, prop_val, text):
                for pattern in [
                    rf'<meta\s+(?:[^>]*?\s)?{prop_attr}=["\']{prop_val}["\']\s+content=["\'](.*?)["\']',
                    rf'<meta\s+(?:[^>]*?\s)?content=["\'](.*?)["\'](?:[^>]*?\s)?{prop_attr}=["\']{prop_val}["\']',
                ]:
                    m = re.search(pattern, text, re.I | re.S)
                    if m:
                        return m.group(1).strip()
                return None
            t = str(html)
            title = extract("property", "og:title", t)
            if not title:
                m = re.search(r'<title[^>]*>(.*?)</title>', t, re.I | re.S)
                title = m.group(1).strip() if m else None
            preview["title"] = title
            preview["description"] = extract("property", "og:description", t) or extract("name", "description", t)
            preview["image"] = extract("property", "og:image", t)
        return {"ok": True, "result": preview, **preview}
    except Exception as e:
        return {"ok": False, "error": str(e)}
