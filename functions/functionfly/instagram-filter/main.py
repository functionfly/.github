FILTERS = {
    "clarendon": {"brightness": 1.1, "contrast": 1.2, "saturation": 1.3},
    "gingham": {"brightness": 1.05, "contrast": 0.9, "saturation": 0.8},
    "moon": {"brightness": 1.1, "contrast": 1.1, "saturation": 0},
    "lark": {"brightness": 1.2, "contrast": 0.9, "saturation": 1.1},
    "reyes": {"brightness": 1.1, "contrast": 0.85, "saturation": 0.75},
    "juno": {"brightness": 1.0, "contrast": 1.1, "saturation": 1.2},
    "slumber": {"brightness": 0.9, "contrast": 0.9, "saturation": 0.6},
    "crema": {"brightness": 1.05, "contrast": 0.9, "saturation": 0.9},
    "ludwig": {"brightness": 1.05, "contrast": 1.05, "saturation": 0.9},
    "aden": {"brightness": 1.2, "contrast": 0.85, "saturation": 0.8},
    "perpetua": {"brightness": 1.1, "contrast": 1.0, "saturation": 1.1},
    "mayfair": {"brightness": 1.1, "contrast": 1.1, "saturation": 1.1},
    "rise": {"brightness": 1.2, "contrast": 0.9, "saturation": 0.9},
    "hudson": {"brightness": 1.2, "contrast": 0.9, "saturation": 1.1},
    "valencia": {"brightness": 1.1, "contrast": 1.1, "saturation": 1.0},
    "xpro2": {"brightness": 1.0, "contrast": 1.3, "saturation": 1.1},
    "willow": {"brightness": 1.1, "contrast": 1.0, "saturation": 0},
    "lofi": {"brightness": 1.0, "contrast": 1.5, "saturation": 1.5},
    "inkwell": {"brightness": 1.0, "contrast": 1.2, "saturation": 0},
    "hefe": {"brightness": 1.0, "contrast": 1.2, "saturation": 1.3},
}


def handler(event):
    image_base64 = event.get("image") if isinstance(event, dict) else None
    filter_name = event.get("filter", "clarendon").lower()
    intensity = float(event.get("intensity", 1.0))
    if not image_base64:
        return {"ok": False, "error": "image (base64) is required"}
    if filter_name not in FILTERS and filter_name != "none":
        return {"ok": False, "error": f"Unknown filter. Available: {', '.join(FILTERS.keys())}"}
    try:
        import base64, io
        from PIL import Image, ImageEnhance
        img_bytes = base64.b64decode(str(image_base64))
        img = Image.open(io.BytesIO(img_bytes)).convert("RGB")
        if filter_name != "none":
            cfg = FILTERS[filter_name]
            def lerp(v, t):
                return 1 + (v - 1) * t
            img = ImageEnhance.Brightness(img).enhance(lerp(cfg["brightness"], intensity))
            img = ImageEnhance.Contrast(img).enhance(lerp(cfg["contrast"], intensity))
            img = ImageEnhance.Color(img).enhance(lerp(cfg.get("saturation", 1), intensity))
        buf = io.BytesIO()
        img.save(buf, format="JPEG", quality=90)
        result = base64.b64encode(buf.getvalue()).decode()
        return {"ok": True, "result": result, "filter": filter_name, "format": "jpeg_base64"}
    except ImportError:
        params = FILTERS.get(filter_name, {})
        return {"ok": True, "result": None, "filter": filter_name, "params": params, "note": "Install Pillow for image output: pip install Pillow"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
