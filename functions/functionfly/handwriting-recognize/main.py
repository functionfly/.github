import hashlib

HANDWRITING_SAMPLES = [
    "Dear John, I hope this letter finds you well.",
    "Shopping list: milk, eggs, bread, butter",
    "Meeting at 3pm tomorrow - don't forget!",
    "The answer is 42",
    "Happy Birthday! Wishing you all the best.",
    "Note to self: call the doctor",
    "Recipe: 2 cups flour, 1 cup sugar, 3 eggs",
    "Phone: 555-0123 Email: test@example.com",
]

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    image_url = event.get("image_url", "")
    image_base64 = event.get("image_base64", "")
    language = event.get("language", "en")
    if not image_url and not image_base64:
        return {"ok": False, "error": "image_url or image_base64 is required"}
    try:
        seed = image_url or image_base64[:50]
        h = hashlib.sha256(seed.encode()).digest()
        text = HANDWRITING_SAMPLES[h[0] % len(HANDWRITING_SAMPLES)]
        return {
            "ok": True,
            "result": text,
            "text": text,
            "confidence": round(0.75 + (h[1] / 255.0) * 0.25, 4),
            "language": language,
            "style": "cursive" if h[2] > 127 else "print",
            "legibility": "high" if h[3] > 200 else ("medium" if h[3] > 100 else "low"),
            "model": "mock-handwriting-v1",
            "note": "Mock handwriting recognition — for production use, integrate Google Vision or Azure Form Recognizer"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
