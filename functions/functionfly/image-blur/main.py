import base64, io

def handler(event):
    image = event.get("image") if isinstance(event, dict) else None
    radius = float(event.get("radius", 2.0))
    fmt = event.get("format", "JPEG")
    if not image:
        return {"ok": False, "error": "image is required"}
    try:
        from PIL import Image, ImageFilter
    except ImportError:
        return {"ok": False, "error": "Pillow is not installed. Install with: pip install Pillow"}
    try:
        img = Image.open(io.BytesIO(base64.b64decode(str(image)))).convert("RGB")
        out = img.filter(ImageFilter.GaussianBlur(radius=radius))
        buf = io.BytesIO(); out.save(buf, format=fmt.upper())
        return {"ok": True, "result": base64.b64encode(buf.getvalue()).decode("utf-8"), "format": fmt, "radius": radius}
    except Exception as e:
        return {"ok": False, "error": str(e)}
