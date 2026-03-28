NOISE_METHODS = {
    "spectral": {"description": "Spectral subtraction", "latency_ms": 10, "quality": "good"},
    "wiener": {"description": "Wiener filter", "latency_ms": 5, "quality": "good"},
    "rnn": {"description": "RNN-based noise suppression", "latency_ms": 20, "quality": "excellent"},
    "deepfilter": {"description": "DeepFilterNet", "latency_ms": 25, "quality": "excellent"},
    "noisereduce": {"description": "Statistical noise reduction", "latency_ms": 50, "quality": "very good"},
    "adaptive": {"description": "Adaptive noise cancellation", "latency_ms": 2, "quality": "good"},
}

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    method = event.get("method", "spectral")
    strength = float(event.get("strength", 0.5))
    audio_url = event.get("audio_url", "")
    audio_base64 = event.get("audio_base64", "")
    if not audio_url and not audio_base64:
        return {"ok": False, "error": "audio_url or audio_base64 is required"}
    try:
        if method not in NOISE_METHODS:
            method = "spectral"
        strength = max(0.0, min(1.0, strength))
        method_info = NOISE_METHODS[method]
        snr_improvement = round(5 + strength * 20, 2)
        return {
            "ok": True,
            "result": {
                "output_url": f"mock://denoised/{method}",
                "snr_improvement_db": snr_improvement,
                "method": method
            },
            "output_url": f"mock://denoised/{method}",
            "method": method,
            "method_info": method_info,
            "strength": strength,
            "snr_improvement_db": snr_improvement,
            "available_methods": list(NOISE_METHODS.keys()),
            "note": "Mock noise reduction — for production use, integrate RNNoise or DeepFilterNet"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
