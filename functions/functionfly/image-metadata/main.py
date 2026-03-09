import base64
import io
import json


def handler(event):
    if isinstance(event, dict):
        image = event.get("image", event.get("data", ""))
    else:
        image = ""

    if not image:
        return {"ok": False, "error": "Input 'image' is required"}

    try:
        raw = base64.b64decode(str(image).strip(), validate=True)
    except Exception as e:
        return {"ok": False, "error": f"Invalid base64 image: {e}"}

    try:
        from PIL import Image
    except ImportError:
        return {"ok": False, "error": "Pillow (PIL) is required; install with: pip install Pillow"}

    try:
        img = Image.open(io.BytesIO(raw))
        width, height = img.size
        out = {
            "ok": True,
            "width": width,
            "height": height,
            "format": img.format or "unknown",
        }
        if hasattr(img, "_getexif") and img._getexif():
            exif = img._getexif()
            if exif:
                try:
                    out["exif"] = {str(k): v for k, v in exif.items()}
                except Exception:
                    out["exif"] = {}
        else:
            out["exif"] = {}
        return out
    except Exception as e:
        return {"ok": False, "error": str(e)}

