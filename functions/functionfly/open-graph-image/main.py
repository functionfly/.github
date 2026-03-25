import re


def handler(event):
    html = event.get("html") if isinstance(event, dict) else None
    url = event.get("url")
    if not html:
        return {"ok": False, "error": "html is required"}
    try:
        text = str(html)
        # Try og:image first
        og_image = None
        for pattern in [
            r'<meta\s+(?:[^>]*?\s)?property=["\']og:image["\']\s+content=["\'](.*?)["\']',
            r'<meta\s+(?:[^>]*?\s)?content=["\'](.*?)["\'](?:[^>]*?\s)?property=["\']og:image["\']',
        ]:
            m = re.search(pattern, text, re.I | re.S)
            if m:
                og_image = m.group(1)
                break
        # Fallback: twitter:image
        twitter_image = None
        for pattern in [
            r'<meta\s+(?:[^>]*?\s)?name=["\']twitter:image["\']\s+content=["\'](.*?)["\']',
            r'<meta\s+(?:[^>]*?\s)?content=["\'](.*?)["\'](?:[^>]*?\s)?name=["\']twitter:image["\']',
        ]:
            m = re.search(pattern, text, re.I | re.S)
            if m:
                twitter_image = m.group(1)
                break
        # Fallback: first img tag
        first_img = None
        m = re.search(r'<img\s[^>]*src=["\']([^"\']+)["\']', text, re.I)
        if m:
            first_img = m.group(1)
        image = og_image or twitter_image or first_img
        return {
            "ok": True,
            "result": image,
            "image_url": image,
            "og_image": og_image,
            "twitter_image": twitter_image,
            "fallback_image": first_img,
            "found": image is not None
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
