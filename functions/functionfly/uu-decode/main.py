import binascii


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    encoding = event.get("encoding", "utf-8")

    if not data:
        return {"ok": False, "error": "data is required"}
    try:
        lines = str(data).splitlines()
        out_bytes = bytearray()
        name = None
        in_data = False
        for line in lines:
            if line.startswith("begin "):
                parts = line.split(" ", 2)
                name = parts[2] if len(parts) > 2 else "file"
                in_data = True
                continue
            if line == "end" or line.startswith("`"):
                in_data = False
                continue
            if in_data and line:
                decoded = binascii.a2b_uu(line + "\n")
                out_bytes.extend(decoded)
        if not name:
            return {"ok": False, "error": "no valid UU header found"}
        try:
            result = out_bytes.decode(encoding)
            is_text = True
        except UnicodeDecodeError:
            import base64
            result = base64.b64encode(bytes(out_bytes)).decode("utf-8")
            is_text = False
        return {"ok": True, "result": result, "name": name, "is_text": is_text}
    except Exception as e:
        return {"ok": False, "error": str(e)}
