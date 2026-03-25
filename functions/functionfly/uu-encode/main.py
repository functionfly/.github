import io
import binascii


def _uu_encode(data: bytes, name: str = "file") -> str:
    lines = []
    lines.append(f"begin 644 {name}")
    i = 0
    while i < len(data):
        chunk = data[i:i+45]
        encoded_len = ((len(chunk) + 2) // 3) * 4
        enc = binascii.b2a_uu(chunk).decode("ascii").rstrip("\n")
        lines.append(enc)
        i += 45
    lines.append("`")
    lines.append("end")
    return "\n".join(lines)


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    name = event.get("name", "file")

    if data is None:
        return {"ok": False, "error": "data is required"}
    try:
        raw = str(data).encode("utf-8")
        result = _uu_encode(raw, str(name))
        return {"ok": True, "result": result}
    except Exception as e:
        return {"ok": False, "error": str(e)}
