def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    comment = event.get("comment", "")
    if not data or not isinstance(data, dict):
        return {"ok": False, "error": "data must be an object"}
    try:
        lines = []
        if comment:
            lines.append(f"# {comment}")
        for k, v in data.items():
            escaped_v = str(v).replace("\\", "\\\\").replace("\n", "\\n").replace("\t", "\\t")
            lines.append(f"{k}={escaped_v}")
        return {"ok": True, "result": "\n".join(lines)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
