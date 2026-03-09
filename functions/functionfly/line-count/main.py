def handler(event):
    if isinstance(event, dict):
        text = event.get("text", event.get("data", ""))
        exclude_empty = event.get("exclude_empty", False)
    else:
        text = str(event) if event is not None else ""
        exclude_empty = False

    if text is None:
        return {"ok": False, "error": "Input 'text' is required"}

    lines = str(text).splitlines()
    if exclude_empty:
        lines = [ln for ln in lines if ln.strip()]
    return {"ok": True, "lines": len(lines)}

