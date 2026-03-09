import re
import urllib.request
import urllib.error


def handler(event):
    """
    Extract Open Graph and Twitter card metadata from a URL's HTML.
    Fetches the page and parses meta tags (og:*, twitter:*).

    Input:
        - url: URL to fetch and extract metadata from (required)

    Returns:
        - ok: True on success
        - metadata: Dict of extracted keys (og_title, og_description, etc.)
        - error: Message if ok is False
    """
    if isinstance(event, dict):
        url = event.get("url", "")
    else:
        url = str(event) if event else ""

    if not url or not str(url).strip():
        return {"ok": False, "error": "Input 'url' is required"}

    url = str(url).strip()
    if not url.startswith(("http://", "https://")):
        url = "https://" + url

    try:
        req = urllib.request.Request(
            url,
            headers={"User-Agent": "FunctionFly-MetadataBot/1.0"},
        )
        with urllib.request.urlopen(req, timeout=10) as resp:
            html = resp.read().decode("utf-8", errors="replace")
    except urllib.error.URLError as e:
        return {"ok": False, "error": f"Failed to fetch URL: {e.reason}"}
    except Exception as e:
        return {"ok": False, "error": str(e)}

    # Parse meta tags: <meta property="og:title" content="..."> or name="twitter:title" content="...">
    meta = {}
    for m in re.finditer(
        r'<meta\s+[^>]*(?:property|name)=["\'](og:|twitter:)([^"\']+)["\'][^>]*content=["\']([^"\']*)["\']',
        html,
        re.IGNORECASE,
    ):
        prefix, key, value = m.group(1), m.group(2), m.group(3)
        key = (prefix + key).replace(":", "_").strip("_")
        if value:
            meta[key] = value

    # Also accept content before name/property
    for m in re.finditer(
        r'<meta\s+[^>]*content=["\']([^"\']*)["\'][^>]*(?:property|name)=["\'](og:|twitter:)([^"\']+)["\']',
        html,
        re.IGNORECASE,
    ):
        value, prefix, key = m.group(1), m.group(2), m.group(3)
        key = (prefix + key).replace(":", "_").strip("_")
        if value and key not in meta:
            meta[key] = value

    return {"ok": True, "metadata": meta, "url": url}
