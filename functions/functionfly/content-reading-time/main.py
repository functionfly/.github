import re


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    wpm = int(event.get("wpm", 238))
    image_time_seconds = float(event.get("image_time_seconds", 12))
    if not text:
        return {"ok": False, "error": "text is required"}
    try:
        stripped = re.sub(r'<[^>]+>', ' ', str(text))
        words = re.findall(r"\b\w+\b", stripped)
        image_count = len(re.findall(r'<img[^>]+>', str(text), re.I))
        text_minutes = len(words) / wpm
        image_minutes = image_count * image_time_seconds / 60
        total_minutes = text_minutes + image_minutes
        minutes = int(total_minutes)
        seconds = int((total_minutes - minutes) * 60)
        formatted = f"{minutes} min read" if minutes >= 1 else f"{max(1, seconds)} sec read"
        return {
            "ok": True,
            "result": round(total_minutes, 2),
            "reading_time_minutes": round(total_minutes, 2),
            "formatted": formatted,
            "word_count": len(words),
            "wpm": wpm
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
