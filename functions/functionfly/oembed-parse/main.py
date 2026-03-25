import json


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    if not data:
        return {"ok": False, "error": "data is required (oEmbed JSON string or object)"}
    try:
        oe = data if isinstance(data, dict) else json.loads(str(data))
        oe_type = oe.get("type", "unknown")
        result = {
            "type": oe_type,
            "title": oe.get("title"),
            "author_name": oe.get("author_name"),
            "author_url": oe.get("author_url"),
            "provider_name": oe.get("provider_name"),
            "provider_url": oe.get("provider_url"),
            "thumbnail_url": oe.get("thumbnail_url"),
            "thumbnail_width": oe.get("thumbnail_width"),
            "thumbnail_height": oe.get("thumbnail_height"),
            "width": oe.get("width"),
            "height": oe.get("height"),
            "html": oe.get("html"),
            "url": oe.get("url"),
            "version": oe.get("version", "1.0"),
        }
        return {"ok": True, "result": result, "oembed": result}
    except Exception as e:
        return {"ok": False, "error": str(e)}
