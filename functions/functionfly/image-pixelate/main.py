import base64, io

def handler(event):
    image = event.get("image") if isinstance(event, dict) else None
    pixel_size = int(event.get("pixel_size", 10))
    fmt = event.get("format", "JPEG")
    if not image:
        return {"ok": False, "error": "image is required"}
    if pixel_size < 1 or pixel_size > 200:
        return {"ok": False, "error": "pixel_size must be 1-200"}
    try:
        from PIL import Image
    except ImportError:
        return {"ok": False, "error": "Pillow is not installed. Install with: pip install Pillow"}
    try:
        img = Image.open(io.BytesIO(base64.b64decode(str(image)))).convert("RGB")
        w, h = img.size
        small = img.resize((max(1, w // pixel_size), max(1, h // pixel_size)), Image.NEAREST)
        out = small.resize((w, h), Image.NEAREST)
        buf = io.BytesIO(); out.save(buf, format=fmt.upper())
        return {"ok": True, "result": base64.b64encode(buf.getvalue()).decode("utf-8"), "format": fmt}
    except Exception as e:
        return {"ok": False, "error": str(e)}
