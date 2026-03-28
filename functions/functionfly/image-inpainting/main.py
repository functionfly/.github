def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    image_url = event.get("image_url", "")
    mask_url = event.get("mask_url", "")
    prompt = event.get("prompt", "")
    method = event.get("method", "diffusion")
    try:
        return {
            "ok": True,
            "result": {
                "output_url": f"mock://inpainted/{method}",
                "method": method,
                "prompt_used": prompt,
                "processing_time_ms": 3000
            },
            "output_url": f"mock://inpainted/{method}",
            "method": method,
            "prompt": prompt,
            "available_methods": ["diffusion", "patch_based", "deep_fill", "lama"],
            "note": "Mock inpainting — for production use, integrate Stable Diffusion inpainting or LaMa"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
