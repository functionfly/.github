import hashlib
import math

PITCH_ALGORITHMS = {
    "yin": {"description": "YIN algorithm", "accuracy": "high", "latency": "low"},
    "pyin": {"description": "Probabilistic YIN", "accuracy": "very high", "latency": "medium"},
    "crepe": {"description": "CREPE neural network", "accuracy": "very high", "latency": "high"},
    "autocorrelation": {"description": "Autocorrelation method", "accuracy": "medium", "latency": "very low"},
    "cepstrum": {"description": "Cepstral analysis", "accuracy": "medium", "latency": "low"},
}

NOTE_NAMES = ["C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"]

def _freq_to_note(freq):
    if freq <= 0:
        return None, None
    semitones = 12 * math.log2(freq / 440.0) + 69
    note_idx = int(round(semitones)) % 12
    octave = int(round(semitones)) // 12 - 1
    return NOTE_NAMES[note_idx], octave

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    audio_url = event.get("audio_url", "")
    audio_base64 = event.get("audio_base64", "")
    algorithm = event.get("algorithm", "yin")
    if not audio_url and not audio_base64:
        return {"ok": False, "error": "audio_url or audio_base64 is required"}
    try:
        seed = audio_url or audio_base64[:50]
        h = hashlib.sha256(seed.encode()).digest()
        if algorithm not in PITCH_ALGORITHMS:
            algorithm = "yin"
        # Generate pitch contour
        base_freq = 80 + (h[0] / 255.0) * 400  # 80-480 Hz
        pitch_frames = []
        for i in range(20):
            freq = base_freq + math.sin(i * 0.5) * 20 + (h[i % 32] / 255.0 - 0.5) * 30
            freq = max(50, min(2000, freq))
            note, octave = _freq_to_note(freq)
            pitch_frames.append({
                "time": round(i * 0.05, 3),
                "frequency": round(freq, 2),
                "note": f"{note}{octave}" if note else None,
                "confidence": round(0.7 + (h[(i+1) % 32] / 255.0) * 0.3, 4)
            })
        mean_freq = sum(f["frequency"] for f in pitch_frames) / len(pitch_frames)
        note, octave = _freq_to_note(mean_freq)
        return {
            "ok": True,
            "result": {"mean_frequency": round(mean_freq, 2), "note": f"{note}{octave}" if note else None},
            "mean_frequency": round(mean_freq, 2),
            "mean_note": f"{note}{octave}" if note else None,
            "pitch_frames": pitch_frames,
            "algorithm": algorithm,
            "algorithm_info": PITCH_ALGORITHMS[algorithm],
            "note": "Mock pitch detection — for production use, integrate librosa or CREPE"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
