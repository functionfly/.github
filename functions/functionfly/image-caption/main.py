import hashlib

CAPTION_TEMPLATES = {
    "descriptive": [
        "A {adj} {subject} {action} in a {setting}.",
        "The image shows {adj} {subject} with {detail}.",
        "A {setting} scene featuring {adj} {subject}.",
        "An outdoor scene with {subject} and {detail}.",
        "A close-up of {adj} {subject} against a {setting} background.",
    ],
    "brief": [
        "{subject} in {setting}",
        "{adj} {subject}",
        "{subject} with {detail}",
        "A {subject}",
        "{setting} with {subject}",
    ]
}

SUBJECTS = ["person","cat","dog","car","building","tree","flower","bird","landscape","group of people","child","woman","man","animal","object"]
ADJECTIVES = ["beautiful","colorful","bright","dark","large","small","old","modern","natural","urban","peaceful","vibrant","detailed","interesting","unique"]
ACTIONS = ["standing","sitting","walking","running","playing","resting","looking","moving","interacting","exploring"]
SETTINGS = ["outdoor","indoor","urban","natural","park","street","room","garden","forest","beach","mountain","city"]
DETAILS = ["colorful background","natural lighting","interesting composition","various objects","people nearby","natural elements","architectural features"]

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}
    image_url = event.get("image_url", "")
    image_base64 = event.get("image_base64", "")
    style = event.get("style", "descriptive")
    if not image_url and not image_base64:
        return {"ok": False, "error": "image_url or image_base64 is required"}
    try:
        seed = image_url or image_base64[:50]
        h = hashlib.sha256(seed.encode()).digest()
        templates = CAPTION_TEMPLATES.get(style, CAPTION_TEMPLATES["descriptive"])
        template = templates[h[0] % len(templates)]
        caption = template.format(
            subject=SUBJECTS[h[1] % len(SUBJECTS)],
            adj=ADJECTIVES[h[2] % len(ADJECTIVES)],
            action=ACTIONS[h[3] % len(ACTIONS)],
            setting=SETTINGS[h[4] % len(SETTINGS)],
            detail=DETAILS[h[5] % len(DETAILS)]
        )
        return {
            "ok": True,
            "result": caption,
            "caption": caption,
            "style": style,
            "confidence": round(0.7 + (h[6] / 255.0) * 0.3, 4),
            "model": "mock-blip-v1",
            "note": "Mock captioning — for production use, integrate BLIP or similar"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
