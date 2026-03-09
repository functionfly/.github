import random
import string


def handler(event):
    if isinstance(event, dict):
        length = event.get("length", 16)
        charset = event.get("charset", "alphanumeric")
    else:
        length = 16
        charset = "alphanumeric"

    try:
        length = int(length)
    except (TypeError, ValueError):
        return {"ok": False, "error": "Invalid 'length'"}

    if length < 1 or length > 1024:
        return {"ok": False, "error": "Length must be between 1 and 1024"}

    if charset == "alphanumeric":
        pool = string.ascii_letters + string.digits
    elif charset == "alpha":
        pool = string.ascii_letters
    elif charset == "numeric":
        pool = string.digits
    elif charset == "hex":
        pool = string.hexdigits.lower()
    else:
        pool = string.ascii_letters + string.digits

    result = "".join(random.choices(pool, k=length))
    return {"ok": True, "result": result}

