import base64, io, math

def handler(event):
    image = event.get("image") if isinstance(event, dict) else None
    if not image:
        return {"ok": False, "error": "image is required"}
    try:
        from PIL import Image
    except ImportError:
        return {"ok": False, "error": "Pillow is not installed. Install with: pip install Pillow"}
    try:
        img = Image.open(io.BytesIO(base64.b64decode(str(image)))).convert("L")
        pixels = list(img.getdata())
        mean = sum(pixels) / len(pixels)
        rms = math.sqrt(sum((p - mean)**2 for p in pixels) / len(pixels))
        return {"ok": True, "result": round(rms, 4), "rms_contrast": round(rms, 4), "mean_luminance": round(mean, 2)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
