import base64
import io


def handler(event):
    image = event.get("image") if isinstance(event, dict) else None
    num_colors = int(event.get("num_colors", 5))
    if not image:
        return {"ok": False, "error": "image is required (Base64-encoded image)"}
    try:
        from PIL import Image
    except ImportError:
        return {"ok": False, "error": "Pillow library is not installed. Install with: pip install Pillow"}
    try:
        img_bytes = base64.b64decode(str(image))
        img = Image.open(io.BytesIO(img_bytes)).convert("RGB")
        img.thumbnail((100, 100))
        pixels = list(img.getdata())
        color_count = {}
        for p in pixels:
            key = (p[0]//32*32, p[1]//32*32, p[2]//32*32)
            color_count[key] = color_count.get(key, 0) + 1
        sorted_colors = sorted(color_count.items(), key=lambda x: x[1], reverse=True)
        dominant = sorted_colors[0][0]
        top_colors = [{"r": c[0][0], "g": c[0][1], "b": c[0][2], "hex": f"#{c[0][0]:02X}{c[0][1]:02X}{c[0][2]:02X}", "count": c[1]} for c in sorted_colors[:num_colors]]
        return {
            "ok": True,
            "result": f"#{dominant[0]:02X}{dominant[1]:02X}{dominant[2]:02X}",
            "r": dominant[0], "g": dominant[1], "b": dominant[2],
            "top_colors": top_colors,
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
