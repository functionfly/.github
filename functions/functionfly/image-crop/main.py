import base64
import io

def handler(event):
    if isinstance(event, dict):
        image = event.get("image", "")
        left = event.get("left")
        top = event.get("top")
        width = event.get("width")
        height = event.get("height")
    else:
        image = left = top = width = height = None
    if not image or left is None or top is None or width is None or height is None:
        return {"ok": False, "error": "image, left, top, width, height are required"}
    try:
        from PIL import Image
    except ImportError:
        return {"ok": False, "error": "Pillow (PIL) is required; pip install Pillow"}
    try:
        raw = base64.b64decode(str(image).strip(), validate=True)
        img = Image.open(io.BytesIO(raw)).convert("RGB")
        box = (int(left), int(top), int(left) + int(width), int(top) + int(height))
        img = img.crop(box)
        buf = io.BytesIO()
        img.save(buf, format="JPEG", quality=90)
        return {"ok": True, "image_base64": base64.b64encode(buf.getvalue()).decode("ascii")}
    except Exception as e:
        return {"ok": False, "error": str(e)}
