import hashlib

SAMPLE_TEXTS = [
    "The quick brown fox jumps over the lazy dog.",
    "Invoice #12345\nDate: 2024-01-15\nTotal: $150.00",
    "STOP\nAll way",
    "Welcome to FunctionFly\nPowered by AI",
    "Name: John Doe\nAddress: 123 Main St\nCity: New York",
    "Chapter 1: Introduction\nThis document describes...",
    "Meeting Notes\n- Action item 1\n- Action item 2\n- Follow up required",
    "Product: Widget Pro\nSKU: WP-001\nPrice: $29.99",
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
        text = SAMPLE_TEXTS[h[0] % len(SAMPLE_TEXTS)]
        lines = text.split("\n")
        words = text.split()
        return {
            "ok": True,
            "result": text,
            "text": text,
            "lines": lines,
            "word_count": len(words),
            "confidence": round(0.85 + (h[1] / 255.0) * 0.15, 4),
            "language": language,
            "bounding_boxes": [{"text": w, "confidence": round(0.8 + (h[i % 32] / 255.0) * 0.2, 4)} for i, w in enumerate(words[:10])],
            "model": "mock-tesseract-v1",
            "note": "Mock OCR — for production use, integrate Tesseract or Google Vision API"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
