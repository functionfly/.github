import hashlib

SAMPLE_TRANSCRIPTS = [
    "Hello, this is a test recording. The audio quality is good.",
    "Welcome to our meeting. Today we will discuss the quarterly results.",
    "The quick brown fox jumps over the lazy dog.",
    "Please leave a message after the beep and we will get back to you.",
    "This is an automated message. Your call is important to us.",
    "Good morning everyone. Let us begin today's presentation.",
    "The weather today is sunny with a high of 75 degrees.",
    "Thank you for calling customer support. How can I help you today?",
]

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    audio_url = event.get("audio_url", "")
    audio_base64 = event.get("audio_base64", "")
    language = event.get("language", "en")
    model = event.get("model", "base")
    if not audio_url and not audio_base64:
        return {"ok": False, "error": "audio_url or audio_base64 is required"}
    try:
        seed = audio_url or audio_base64[:50]
        h = hashlib.sha256(seed.encode()).digest()
        transcript = SAMPLE_TRANSCRIPTS[h[0] % len(SAMPLE_TRANSCRIPTS)]
        words = transcript.split()
        word_timestamps = []
        t = 0.0
        for word in words:
            duration = 0.3 + (len(word) * 0.05)
            word_timestamps.append({"word": word, "start": round(t, 2), "end": round(t + duration, 2), "confidence": round(0.85 + (h[len(word) % 32] / 255.0) * 0.15, 4)})
            t += duration + 0.1
        return {
            "ok": True,
            "result": transcript,
            "transcript": transcript,
            "language": language,
            "model": f"mock-whisper-{model}",
            "duration_seconds": round(t, 2),
            "word_count": len(words),
            "confidence": round(0.9 + (h[1] / 255.0) * 0.1, 4),
            "words": word_timestamps,
            "note": "Mock transcription — for production use, integrate Whisper or similar"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
