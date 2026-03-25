import base64, io

def handler(event):
    image = event.get("image") if isinstance(event, dict) else None
    colors = int(event.get("colors", 8))
    fmt = event.get("format", "PNG")
    if not image:
        return {"ok": False, "error": "image is required"}
    try:
        from PIL import Image
    except ImportError:
        return {"ok": False, "error": "Pillow is not installed. Install with: pip install Pillow"}
    try:
        img = Image.open(io.BytesIO(base64.b64decode(str(image)))).convert("RGB")
        quantized = img.quantize(colors=colors, method=Image.MEDIANCUT)
        palette = quantized.getpalette()[:colors*3]
        color_list = [{"r": palette[i*3], "g": palette[i*3+1], "b": palette[i*3+2],
                       "hex": f"#{palette[i*3]:02X}{palette[i*3+1]:02X}{palette[i*3+2]:02X}"}
                      for i in range(colors)]
        out = quantized.convert("RGB")
        buf = io.BytesIO(); out.save(buf, format=fmt.upper())
        return {"ok": True, "result": base64.b64encode(buf.getvalue()).decode("utf-8"), "palette": color_list, "format": fmt}
    except Exception as e:
        return {"ok": False, "error": str(e)}
