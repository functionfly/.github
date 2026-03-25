import base64, io

def handler(event):
    image = event.get("image") if isinstance(event, dict) else None
    text = event.get("text", "Watermark")
    opacity = int(event.get("opacity", 128))
    position = event.get("position", "bottom-right")
    font_size = int(event.get("font_size", 24))
    fmt = event.get("format", "PNG")
    if not image:
        return {"ok": False, "error": "image is required"}
    try:
        from PIL import Image, ImageDraw, ImageFont
    except ImportError:
        return {"ok": False, "error": "Pillow is not installed. Install with: pip install Pillow"}
    try:
        img = Image.open(io.BytesIO(base64.b64decode(str(image)))).convert("RGBA")
        overlay = Image.new("RGBA", img.size, (255,255,255,0))
        draw = ImageDraw.Draw(overlay)
        try:
            font = ImageFont.truetype("/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf", font_size)
        except Exception:
            font = ImageFont.load_default()
        bbox = draw.textbbox((0,0), str(text), font=font)
        tw, th = bbox[2]-bbox[0], bbox[3]-bbox[1]
        w, h = img.size; margin = 20
        if position == "bottom-right":
            xy = (w-tw-margin, h-th-margin)
        elif position == "bottom-left":
            xy = (margin, h-th-margin)
        elif position == "top-right":
            xy = (w-tw-margin, margin)
        elif position == "top-left":
            xy = (margin, margin)
        else:
            xy = ((w-tw)//2, (h-th)//2)
        draw.text(xy, str(text), font=font, fill=(255,255,255,opacity))
        out = Image.alpha_composite(img, overlay).convert("RGB")
        buf = io.BytesIO(); out.save(buf, format=fmt.upper())
        return {"ok": True, "result": base64.b64encode(buf.getvalue()).decode("utf-8"), "format": fmt}
    except Exception as e:
        return {"ok": False, "error": str(e)}
