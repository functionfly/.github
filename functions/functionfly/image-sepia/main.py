import base64, io

def handler(event):
    image = event.get("image") if isinstance(event, dict) else None
    intensity = float(event.get("intensity", 1.0))
    fmt = event.get("format", "JPEG")
    if not image:
        return {"ok": False, "error": "image is required"}
    try:
        from PIL import Image
    except ImportError:
        return {"ok": False, "error": "Pillow is not installed. Install with: pip install Pillow"}
    try:
        img = Image.open(io.BytesIO(base64.b64decode(str(image)))).convert("RGB")
        pixels = img.load(); w, h = img.size
        for y in range(h):
            for x in range(w):
                r, g, b = pixels[x, y]
                nr = min(255, round(r*0.393 + g*0.769 + b*0.189))
                ng = min(255, round(r*0.349 + g*0.686 + b*0.168))
                nb = min(255, round(r*0.272 + g*0.534 + b*0.131))
                # Blend by intensity
                pixels[x, y] = (round(r + (nr-r)*intensity), round(g + (ng-g)*intensity), round(b + (nb-b)*intensity))
        buf = io.BytesIO(); img.save(buf, format=fmt.upper())
        return {"ok": True, "result": base64.b64encode(buf.getvalue()).decode("utf-8"), "format": fmt}
    except Exception as e:
        return {"ok": False, "error": str(e)}
