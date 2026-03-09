import base64
import io


def handler(event):
    if isinstance(event, dict):
        image = event.get("image", event.get("data", ""))
        quality = event.get("quality", 85)
        max_width = event.get("max_width")
        max_height = event.get("max_height")
    else:
        image = ""
        quality = 85
        max_width = max_height = None

    if not image:
        return {"ok": False, "error": "Input 'image' is required"}

    try:
        quality = max(1, min(95, int(quality)))
    except (TypeError, ValueError):
        quality = 85

    try:
        from PIL import Image
    except ImportError:
        return {"ok": False, "error": "Pillow (PIL) is required; install with: pip install Pillow"}

    try:
        raw = base64.b64decode(str(image).strip(), validate=True)
        bytes_original = len(raw)
        img = Image.open(io.BytesIO(raw)).convert("RGB")
        if max_width or max_height:
            w, h = img.size
            mw = max_width or w
            mh = max_height or h
            img.thumbnail((mw, mh), Image.Resampling.LANCZOS)
        buf = io.BytesIO()
        img.save(buf, format="JPEG", quality=quality, optimize=True)
        out_b64 = base64.b64encode(buf.getvalue()).decode("ascii")
        return {
            "ok": True,
            "image_base64": out_b64,
            "bytes_original": bytes_original,
            "bytes_compressed": len(buf.getvalue()),
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}

