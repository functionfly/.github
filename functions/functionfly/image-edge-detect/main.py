import base64, io

def handler(event):
    image = event.get("image") if isinstance(event, dict) else None
    fmt = event.get("format", "PNG")
    mode = event.get("mode", "edges")
    if not image:
        return {"ok": False, "error": "image is required"}
    try:
        from PIL import Image, ImageFilter
    except ImportError:
        return {"ok": False, "error": "Pillow is not installed. Install with: pip install Pillow"}
    try:
        img = Image.open(io.BytesIO(base64.b64decode(str(image)))).convert("RGB")
        if mode == "edges":
            out = img.filter(ImageFilter.FIND_EDGES)
        elif mode == "emboss":
            out = img.filter(ImageFilter.EMBOSS)
        elif mode == "contour":
            out = img.filter(ImageFilter.CONTOUR)
        else:
            out = img.filter(ImageFilter.FIND_EDGES)
        buf = io.BytesIO(); out.save(buf, format=fmt.upper())
        return {"ok": True, "result": base64.b64encode(buf.getvalue()).decode("utf-8"), "format": fmt, "mode": mode}
    except Exception as e:
        return {"ok": False, "error": str(e)}
