import re

VOICES = {
    "neutral": {"gender": "neutral", "age": "adult", "style": "standard"},
    "male": {"gender": "male", "age": "adult", "style": "standard"},
    "female": {"gender": "female", "age": "adult", "style": "standard"},
    "child": {"gender": "neutral", "age": "child", "style": "standard"},
    "elderly": {"gender": "neutral", "age": "elderly", "style": "standard"},
    "narrator": {"gender": "male", "age": "adult", "style": "narrative"},
    "assistant": {"gender": "female", "age": "adult", "style": "conversational"},
    "news": {"gender": "neutral", "age": "adult", "style": "broadcast"},
}

LANGUAGES = {
    "en": "English", "es": "Spanish", "fr": "French", "de": "German",
    "it": "Italian", "pt": "Portuguese", "ja": "Japanese", "zh": "Chinese",
    "ko": "Korean", "ar": "Arabic", "ru": "Russian", "hi": "Hindi",
}

def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    if not text or not isinstance(text, str):
        return {"ok": False, "error": "text (string) is required"}
    try:
        voice = event.get("voice", "neutral")
        language = event.get("language", "en")
        speed = float(event.get("speed", 1.0))
        pitch = float(event.get("pitch", 1.0))
        if voice not in VOICES:
            voice = "neutral"
        if language not in LANGUAGES:
            language = "en"
        speed = max(0.5, min(2.0, speed))
        pitch = max(0.5, min(2.0, pitch))
        words = len(text.split())
        # Estimate duration: average 150 words per minute at speed 1.0
        duration_seconds = round((words / 150) * 60 / speed, 2)
        phoneme_count = sum(1 for c in text.lower() if c in "aeiou") * 2 + len(re.findall(r"[bcdfghjklmnpqrstvwxyz]", text.lower()))
        return {
            "ok": True,
            "result": {
                "audio_url": f"mock://tts/{language}/{voice}/{hash(text) % 100000}",
                "duration_seconds": duration_seconds,
                "format": "mp3"
            },
            "audio_url": f"mock://tts/{language}/{voice}/{hash(text) % 100000}",
            "duration_seconds": duration_seconds,
            "voice": voice,
            "voice_info": VOICES[voice],
            "language": language,
            "language_name": LANGUAGES[language],
            "speed": speed,
            "pitch": pitch,
            "word_count": words,
            "phoneme_count": phoneme_count,
            "format": "mp3",
            "sample_rate": 22050,
            "available_voices": list(VOICES.keys()),
            "note": "Mock TTS — for production use, integrate ElevenLabs, Google TTS, or similar"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
