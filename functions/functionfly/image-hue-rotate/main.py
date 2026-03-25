import base64, io

def handler(event):
    image = event.get("image") if isinstance(event, dict) else None
    degrees = float(event.get("degrees", 90))
    fmt = event.get("format", "PNG")
    if not image:
        return {"ok": False, "error": "image is required"}
    try:
        from PIL import Image
    except ImportError:
        return {"ok": False, "error": "Pillow is not installed. Install with: pip install Pillow"}
    try:
        img = Image.open(io.BytesIO(base64.b64decode(str(image)))).convert("HSV")
        pixels = img.load()
        w, h = img.size
        for y in range(h):
            for x in range(w):
                hv, s, v = pixels[x, y]
                new_h = (hv + int(degrees / 360 * 255)) % 256
                pixels[x, y] = (new_h, s, v)
        out = img.convert("RGB")
        buf = io.BytesIO(); out.save(buf, format=fmt.upper())
        return {"ok": True, "result": base64.b64encode(buf.getvalue()).decode("utf-8"), "format": fmt}
    except Exception as e:
        return {"ok": False, "error": str(e)}
