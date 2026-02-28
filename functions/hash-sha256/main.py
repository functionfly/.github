import hashlib


def handler(event):
    try:
        if isinstance(event, dict):
            text = event.get("text", "")
        else:
            text = str(event) if event is not None else ""

        if text is None or (isinstance(text, str) and not text and text != ""):
            return {"ok": False, "error": "Missing required field: text"}

        if not isinstance(text, str):
            text = str(text)

        hash_hex = hashlib.sha256(text.encode("utf-8")).hexdigest()
        return {"ok": True, "hash": hash_hex}
    except Exception as e:
        return {"ok": False, "error": str(e)}
