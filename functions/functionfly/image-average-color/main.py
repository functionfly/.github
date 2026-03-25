import base64, io

def handler(event):
    image = event.get("image") if isinstance(event, dict) else None
    if not image:
        return {"ok": False, "error": "image is required"}
    try:
        from PIL import Image
    except ImportError:
        return {"ok": False, "error": "Pillow is not installed. Install with: pip install Pillow"}
    try:
        img = Image.open(io.BytesIO(base64.b64decode(str(image)))).convert("RGB")
        pixels = list(img.getdata())
        n = len(pixels)
        r = round(sum(p[0] for p in pixels) / n)
        g = round(sum(p[1] for p in pixels) / n)
        b = round(sum(p[2] for p in pixels) / n)
        return {"ok": True, "result": f"#{r:02X}{g:02X}{b:02X}", "r": r, "g": g, "b": b}
    except Exception as e:
        return {"ok": False, "error": str(e)}
