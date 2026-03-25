import email.header


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None

    if data is None:
        return {"ok": False, "error": "data is required"}
    try:
        decoded_parts = email.header.decode_header(str(data))
        parts = []
        full_text = ""
        for part, charset in decoded_parts:
            if isinstance(part, bytes):
                text = part.decode(charset or "utf-8", errors="replace")
            else:
                text = part
            parts.append({"text": text, "charset": charset or "ascii"})
            full_text += text
        return {"ok": True, "result": full_text, "parts": parts}
    except Exception as e:
        return {"ok": False, "error": str(e)}
