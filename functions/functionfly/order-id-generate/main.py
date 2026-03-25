import os, time, random, string


def handler(event):
    prefix = event.get("prefix", "ORD") if isinstance(event, dict) else "ORD"
    length = int(event.get("length", 10))
    style = event.get("style", "alphanumeric")
    try:
        ts = str(int(time.time() * 1000))[-6:]
        if style == "numeric":
            rand = "".join(random.choices(string.digits, k=max(4, length - len(prefix) - 1)))
            order_id = f"{prefix}-{rand}"
        elif style == "uuid-short":
            rand_hex = os.urandom(6).hex().upper()
            order_id = f"{prefix}-{ts}-{rand_hex}"
        else:
            chars = string.ascii_uppercase + string.digits
            rand = "".join(random.choices(chars, k=max(6, length)))
            order_id = f"{prefix}-{rand}"
        return {"ok": True, "result": order_id, "order_id": order_id, "prefix": prefix}
    except Exception as e:
        return {"ok": False, "error": str(e)}
