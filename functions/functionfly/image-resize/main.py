import base64
import io


def handler(event):
    if isinstance(event, dict):
        image = event.get("image", event.get("data", ""))
        width = event.get("width")
        height = event.get("height")
        fit = (event.get("fit") or "fill").lower()
    else:
        image, width, height, fit = "", None, None, "fill"

    if not image:
        return {"ok": False, "error": "Input 'image' is required"}
    if width is None and height is None:
        return {"ok": False, "error": "At least one of width or height is required"}

    try:
        from PIL import Image
    except ImportError:
        return {"ok": False, "error": "Pillow (PIL) is required; install with: pip install Pillow"}

    try:
        raw = base64.b64decode(str(image).strip(), validate=True)
        img = Image.open(io.BytesIO(raw)).convert("RGB")
        w, h = img.size
        tw = int(width) if width is not None else None
        th = int(height) if height is not None else None
        if tw is not None and th is None:
            th = int(h * tw / w) if w else tw
        elif th is not None and tw is None:
            tw = int(w * th / h) if h else th
        else:
            tw = tw or w
            th = th or h

        if fit == "cover":
            r = max(tw / w, th / h) if w and h else 1
            nw, nh = int(w * r), int(h * r)
            img = img.resize((nw, nh), Image.Resampling.LANCZOS)
            left = (nw - tw) // 2
            top = (nh - th) // 2
            img = img.crop((left, top, left + tw, top + th))
        elif fit == "contain":
            r = min(tw / w, th / h) if w and h else 1
            nw, nh = int(w * r), int(h * r)
            img = img.resize((nw, nh), Image.Resampling.LANCZOS)
        else:
            img = img.resize((tw, th), Image.Resampling.LANCZOS)

        buf = io.BytesIO()
        img.save(buf, format="JPEG", quality=90)
        out_b64 = base64.b64encode(buf.getvalue()).decode("ascii")
        return {"ok": True, "image_base64": out_b64, "width": img.size[0], "height": img.size[1]}
    except Exception as e:
        return {"ok": False, "error": str(e)}
