import base64
import io

def handler(event):
    if isinstance(event, dict):
        image = event.get("image", "")
        degrees = event.get("degrees")
        expand = event.get("expand", True)
    else:
        image = degrees = None
        expand = True
    if not image or degrees is None:
        return {"ok": False, "error": "image and degrees are required"}
    try:
        from PIL import Image
    except ImportError:
        return {"ok": False, "error": "Pillow (PIL) is required; pip install Pillow"}
    try:
        raw = base64.b64decode(str(image).strip(), validate=True)
        img = Image.open(io.BytesIO(raw)).convert("RGB")
        angle = float(degrees)
        img = img.rotate(-angle, expand=bool(expand), resample=Image.Resampling.BICUBIC)
        buf = io.BytesIO()
        img.save(buf, format="JPEG", quality=90)
        return {"ok": True, "image_base64": base64.b64encode(buf.getvalue()).decode("ascii")}
    except Exception as e:
        return {"ok": False, "error": str(e)}
