import base64, io

def handler(event):
    image = event.get("image") if isinstance(event, dict) else None
    colors = int(event.get("colors", 16))
    fmt = event.get("format", "PNG")
    if not image:
        return {"ok": False, "error": "image is required"}
    if colors < 2 or colors > 256:
        return {"ok": False, "error": "colors must be 2-256"}
    try:
        from PIL import Image
    except ImportError:
        return {"ok": False, "error": "Pillow is not installed. Install with: pip install Pillow"}
    try:
        img = Image.open(io.BytesIO(base64.b64decode(str(image)))).convert("RGB")
        dithered = img.convert("P", palette=Image.ADAPTIVE, colors=colors, dither=Image.FLOYDSTEINBERG)
        out = dithered.convert("RGB")
        buf = io.BytesIO(); out.save(buf, format=fmt.upper())
        return {"ok": True, "result": base64.b64encode(buf.getvalue()).decode("utf-8"), "format": fmt, "colors": colors}
    except Exception as e:
        return {"ok": False, "error": str(e)}
