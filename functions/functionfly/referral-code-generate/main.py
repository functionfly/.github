import random
import string
import hashlib


def handler(event):
    user_id = event.get("user_id") if isinstance(event, dict) else None
    length = int(event.get("length", 8))
    prefix = event.get("prefix", "")
    style = event.get("style", "random")
    if not user_id and style == "hash":
        return {"ok": False, "error": "user_id is required for hash style"}
    try:
        length = max(4, min(length, 32))
        if style == "hash" and user_id:
            h = hashlib.sha256(str(user_id).encode()).hexdigest()
            code = h[:length].upper()
        elif style == "numeric":
            code = ''.join(random.choices(string.digits, k=length))
        elif style == "alpha":
            chars = string.ascii_uppercase
            code = ''.join(random.choices(chars, k=length))
        else:
            chars = string.ascii_uppercase + string.digits
            code = ''.join(random.choices(chars, k=length))
        if prefix:
            code = f"{str(prefix).upper()}{code}"
        return {
            "ok": True,
            "result": code,
            "referral_code": code,
            "length": len(code),
            "style": style,
            "user_id": str(user_id) if user_id else None
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
