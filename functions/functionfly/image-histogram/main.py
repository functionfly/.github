import base64, io

def handler(event):
    image = event.get("image") if isinstance(event, dict) else None
    channel = event.get("channel", "all")
    if not image:
        return {"ok": False, "error": "image is required"}
    try:
        from PIL import Image
    except ImportError:
        return {"ok": False, "error": "Pillow is not installed. Install with: pip install Pillow"}
    try:
        img = Image.open(io.BytesIO(base64.b64decode(str(image)))).convert("RGB")
        hist = img.histogram()
        r_hist = hist[0:256]
        g_hist = hist[256:512]
        b_hist = hist[512:768]
        if channel == "r":
            return {"ok": True, "result": r_hist, "channel": "r"}
        elif channel == "g":
            return {"ok": True, "result": g_hist, "channel": "g"}
        elif channel == "b":
            return {"ok": True, "result": b_hist, "channel": "b"}
        else:
            return {"ok": True, "result": {"r": r_hist, "g": g_hist, "b": b_hist}, "channel": "all"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
