import base64, io

def handler(event):
    image = event.get("image") if isinstance(event, dict) else None
    threshold = int(event.get("threshold", 128))
    fmt = event.get("format", "PNG")
    invert = event.get("invert", False)
    if not image:
        return {"ok": False, "error": "image is required"}
    if not 0 <= threshold <= 255:
        return {"ok": False, "error": "threshold must be 0-255"}
    try:
        from PIL import Image
    except ImportError:
        return {"ok": False, "error": "Pillow is not installed. Install with: pip install Pillow"}
    try:
        img = Image.open(io.BytesIO(base64.b64decode(str(image)))).convert("L")
        pixels = img.load(); w, h = img.size
        for y in range(h):
            for x in range(w):
                v = pixels[x, y]
                pixels[x, y] = 255 if (v >= threshold) != invert else 0
        buf = io.BytesIO(); img.save(buf, format=fmt.upper())
        return {"ok": True, "result": base64.b64encode(buf.getvalue()).decode("utf-8"), "format": fmt, "threshold": threshold}
    except Exception as e:
        return {"ok": False, "error": str(e)}
