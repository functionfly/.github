import hashlib
import math

BEAT_ALGORITHMS = {
    "onset": {"description": "Onset detection", "accuracy": "high"},
    "tempogram": {"description": "Tempogram analysis", "accuracy": "very high"},
    "beat_tracker": {"description": "Dynamic programming beat tracker", "accuracy": "high"},
    "madmom": {"description": "madmom neural network", "accuracy": "very high"},
}

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    audio_url = event.get("audio_url", "")
    audio_base64 = event.get("audio_base64", "")
    algorithm = event.get("algorithm", "onset")
    if not audio_url and not audio_base64:
        return {"ok": False, "error": "audio_url or audio_base64 is required"}
    try:
        seed = audio_url or audio_base64[:50]
        h = hashlib.sha256(seed.encode()).digest()
        if algorithm not in BEAT_ALGORITHMS:
            algorithm = "onset"
        # Generate BPM and beat times
        bpm = 60 + (h[0] / 255.0) * 120  # 60-180 BPM
        beat_interval = 60.0 / bpm
        beats = []
        t = 0.0
        for i in range(32):
            jitter = (h[i % 32] / 255.0 - 0.5) * 0.02  # Small timing jitter
            beats.append({
                "time": round(t + jitter, 4),
                "strength": round(0.5 + (h[(i+1) % 32] / 255.0) * 0.5, 4),
                "beat_number": i + 1,
                "downbeat": (i % 4 == 0)
            })
            t += beat_interval
        # Time signature detection
        time_sig_num = [3, 4, 6][h[1] % 3]
        return {
            "ok": True,
            "result": {"bpm": round(bpm, 2), "beats": beats},
            "bpm": round(bpm, 2),
            "bpm_confidence": round(0.8 + (h[2] / 255.0) * 0.2, 4),
            "beats": beats,
            "beat_count": len(beats),
            "time_signature": f"{time_sig_num}/4",
            "duration_seconds": round(t, 2),
            "algorithm": algorithm,
            "algorithm_info": BEAT_ALGORITHMS[algorithm],
            "note": "Mock beat detection — for production use, integrate librosa or madmom"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
