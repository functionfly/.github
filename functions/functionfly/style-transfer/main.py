STYLE_PARAMS = {
    "impressionist": {"brush_size": 8, "color_palette": "warm", "texture": "painterly", "abstraction": 0.6},
    "cubist": {"brush_size": 12, "color_palette": "geometric", "texture": "angular", "abstraction": 0.8},
    "watercolor": {"brush_size": 15, "color_palette": "soft", "texture": "fluid", "abstraction": 0.4},
    "oil_painting": {"brush_size": 6, "color_palette": "rich", "texture": "thick", "abstraction": 0.3},
    "sketch": {"brush_size": 2, "color_palette": "grayscale", "texture": "line", "abstraction": 0.5},
    "anime": {"brush_size": 3, "color_palette": "vivid", "texture": "flat", "abstraction": 0.4},
    "pop_art": {"brush_size": 5, "color_palette": "bold", "texture": "flat", "abstraction": 0.5},
    "abstract": {"brush_size": 20, "color_palette": "varied", "texture": "expressive", "abstraction": 0.9},
}

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    style = event.get("style", "impressionist")
    strength = float(event.get("strength", 0.7))
    content_url = event.get("content_image_url", "")
    style_url = event.get("style_image_url", "")
    try:
        params = STYLE_PARAMS.get(style, STYLE_PARAMS["impressionist"])
        return {
            "ok": True,
            "result": {
                "style": style,
                "strength": strength,
                "parameters": params,
                "output_url": f"mock://style-transfer/{style}/{strength}",
                "processing_time_ms": 2500
            },
            "style": style,
            "strength": strength,
            "parameters": params,
            "available_styles": list(STYLE_PARAMS.keys()),
            "note": "Mock style transfer — for production use, integrate neural style transfer model"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
