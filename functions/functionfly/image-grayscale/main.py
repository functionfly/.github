import base64
import io


def handler(event):
    if isinstance(event, dict):
        image = event.get("image", event.get("data", ""))
    else:
        image = ""

    if not image:
        return {"ok": False, "error": "Input 'image' is required"}

    try:
        from PIL import Image
    except ImportError:
        return {"ok": False, "error": "Pillow (PIL) is required; install with: pip install Pillow"}

    try:
        raw = base64.b64decode(str(image).strip(), validate=True)
        img = Image.open(io.BytesIO(raw)).convert("L")
        buf = io.BytesIO()
        img.save(buf, format="PNG")
        out_b64 = base64.b64encode(buf.getvalue()).decode("ascii")
        return {"ok": True, "image_base64": out_b64}
    except Exception as e:
        return {"ok": False, "error": str(e)}

