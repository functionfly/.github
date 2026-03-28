import xml.etree.ElementTree as ET


def handler(event):
    """Validate a sitemap XML and extract URLs."""
    try:
        content = event.get("content")
        if not content:
            return {"ok": False, "error": "content is required"}

        errors = []
        urls = []

        try:
            root = ET.fromstring(content)
        except ET.ParseError as e:
            return {"ok": True, "valid": False, "errors": [str(e)], "url_count": 0, "urls": []}

        # Handle namespace
        ns = ""
        if root.tag.startswith("{"):
            ns = root.tag.split("}")[0] + "}"

        tag = root.tag.replace(ns, "")

        if tag == "urlset":
            for url_elem in root.findall(f"{ns}url"):
                loc = url_elem.find(f"{ns}loc")
                lastmod = url_elem.find(f"{ns}lastmod")
                changefreq = url_elem.find(f"{ns}changefreq")
                priority = url_elem.find(f"{ns}priority")

                if loc is None or not loc.text:
                    errors.append("URL entry missing <loc> element")
                    continue

                entry = {"loc": loc.text.strip()}
                if lastmod is not None and lastmod.text:
                    entry["lastmod"] = lastmod.text.strip()
                if changefreq is not None and changefreq.text:
                    entry["changefreq"] = changefreq.text.strip()
                if priority is not None and priority.text:
                    entry["priority"] = priority.text.strip()
                urls.append(entry)

        elif tag == "sitemapindex":
            for sitemap_elem in root.findall(f"{ns}sitemap"):
                loc = sitemap_elem.find(f"{ns}loc")
                if loc is not None and loc.text:
                    urls.append({"loc": loc.text.strip(), "type": "sitemap"})
        else:
            errors.append(f"Unknown root element: {tag}")

        return {
            "ok": True,
            "valid": len(errors) == 0,
            "url_count": len(urls),
            "urls": urls[:100],  # Limit to 100
            "errors": errors,
            "type": tag,
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
