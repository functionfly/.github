import base64, io

def handler(event):
    image = event.get("image") if isinstance(event, dict) else None
    if not image:
        return {"ok": False, "error": "image is required"}
    try:
        from rembg import remove
        from PIL import Image
    except ImportError:
        try:
            from PIL import Image
            img = Image.open(io.BytesIO(base64.b64decode(str(image)))).convert("RGBA")
            r,g,b,a = img.split()
            bg_r,bg_g,bg_b = 255,255,255
            pixels = list(img.getdata())
            new_pixels = []
            for pr,pg,pb,pa in pixels:
                dist = ((pr-bg_r)**2+(pg-bg_g)**2+(pb-bg_b)**2)**0.5
                new_a = 0 if dist < 80 else pa
                new_pixels.append((pr,pg,pb,new_a))
            img.putdata(new_pixels)
            buf = io.BytesIO(); img.save(buf, format="PNG")
            return {"ok": True, "result": base64.b64encode(buf.getvalue()).decode("utf-8"), "format": "PNG",
                    "note": "rembg not installed, used simple threshold removal"}
        except ImportError:
            return {"ok": False, "error": "Pillow is not installed. Install with: pip install Pillow rembg"}
    try:
        img_bytes = base64.b64decode(str(image))
        output = remove(img_bytes)
        return {"ok": True, "result": base64.b64encode(output).decode("utf-8"), "format": "PNG"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
