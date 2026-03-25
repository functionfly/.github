import os, hashlib, string


def handler(event):
    url = event.get("url") if isinstance(event, dict) else None
    length = int(event.get("length", 7))
    base_url = event.get("base_url", "https://sht.fly/")
    custom_code = event.get("custom_code")
    if not url:
        return {"ok": False, "error": "url is required"}
    try:
        if custom_code:
            code = str(custom_code)
        else:
            seed = hashlib.sha256(f"{url}{os.urandom(4).hex()}".encode()).hexdigest()
            chars = string.ascii_letters + string.digits
            code = "".join(chars[int(seed[i:i+2], 16) % len(chars)] for i in range(0, length * 2, 2))[:length]
        short_url = f"{base_url.rstrip('/')}/{code}"
        return {"ok": True, "result": short_url, "short_url": short_url, "code": code, "original_url": str(url)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
