import base64
import io

def handler(event):
    if isinstance(event, dict):
        image = event.get("image", "")
        xc = event.get("x_components", 4)
        yc = event.get("y_components", 3)
    else:
        image, xc, yc = "", 4, 3
    if not image:
        return {"ok": False, "error": "image is required"}
    try:
        import blurhash
        from PIL import Image
    except ImportError:
        return {"ok": False, "error": "blurhash and Pillow required; pip install blurhash Pillow"}
    try:
        raw = base64.b64decode(str(image).strip(), validate=True)
        img = Image.open(io.BytesIO(raw)).convert("RGB")
        import numpy as np
        arr = np.array(img)
        xc = max(1, min(9, int(xc)))
        yc = max(1, min(9, int(yc)))
        hsh = blurhash.encode(arr, components_x=xc, components_y=yc)
        return {"ok": True, "blurhash": hsh}
    except Exception as e:
        return {"ok": False, "error": str(e)}
