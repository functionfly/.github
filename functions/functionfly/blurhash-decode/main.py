import base64
import io

def handler(event):
    if isinstance(event, dict):
        hsh = event.get("blurhash", "")
        w = event.get("width", 32)
        h = event.get("height", 32)
        punch = event.get("punch", 1.0)
    else:
        hsh, w, h, punch = "", 32, 32, 1.0
    if not hsh:
        return {"ok": False, "error": "blurhash is required"}
    try:
        import blurhash
        from PIL import Image
    except ImportError:
        return {"ok": False, "error": "blurhash and Pillow required; pip install blurhash Pillow"}
    try:
        w = max(1, min(256, int(w)))
        h = max(1, min(256, int(h)))
        punch = float(punch)
        pixels = blurhash.decode(hsh, width=w, height=h, punch=punch)
        import numpy as np
        arr = (np.clip(pixels, 0, 1) * 255).astype("uint8")
        img = Image.fromarray(arr)
        buf = io.BytesIO()
        img.save(buf, format="PNG")
        return {"ok": True, "image_base64": base64.b64encode(buf.getvalue()).decode("ascii")}
    except Exception as e:
        return {"ok": False, "error": str(e)}
