DENOISE_METHODS = {
    "gaussian": {"kernel_size": 5, "sigma": 1.0, "description": "Gaussian blur denoising"},
    "median": {"kernel_size": 3, "description": "Median filter denoising"},
    "bilateral": {"d": 9, "sigma_color": 75, "sigma_space": 75, "description": "Bilateral filter (edge-preserving)"},
    "nlm": {"h": 10, "template_window": 7, "search_window": 21, "description": "Non-local means denoising"},
    "wavelet": {"wavelet": "db1", "level": 2, "description": "Wavelet-based denoising"},
    "dncnn": {"depth": 17, "description": "DnCNN deep learning denoising"},
}

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    method = event.get("method", "gaussian")
    strength = float(event.get("strength", 0.5))
    image_url = event.get("image_url", "")
    image_base64 = event.get("image_base64", "")
    if not image_url and not image_base64:
        return {"ok": False, "error": "image_url or image_base64 is required"}
    try:
        if method not in DENOISE_METHODS:
            method = "gaussian"
        params = DENOISE_METHODS[method].copy()
        params["strength"] = strength
        return {
            "ok": True,
            "result": {"method": method, "parameters": params, "output_url": f"mock://denoised/{method}"},
            "method": method,
            "parameters": params,
            "estimated_psnr_improvement": round(2 + strength * 8, 2),
            "available_methods": list(DENOISE_METHODS.keys()),
            "note": "Mock denoising — for production use, integrate OpenCV or a denoising model"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
