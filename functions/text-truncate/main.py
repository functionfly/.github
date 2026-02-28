def handler(event):
    try:
        if isinstance(event, dict):
            text = event.get("text", "")
            max_length = event.get("max_length", 100)
            suffix = event.get("suffix", "...")
        else:
            text = str(event) if event is not None else ""
            max_length = 100
            suffix = "..."

        if text is None:
            return {"ok": False, "error": "Missing required field: text"}

        if not isinstance(text, str):
            text = str(text)

        try:
            max_length = int(max_length)
            if max_length < 0:
                max_length = 0
        except (TypeError, ValueError):
            return {"ok": False, "error": "max_length must be an integer"}

        if not isinstance(suffix, str):
            suffix = str(suffix)

        if len(text) <= max_length:
            result = text
        else:
            truncate_at = max_length - len(suffix)
            if truncate_at < 0:
                truncate_at = 0
            result = text[:truncate_at] + suffix

        return {"ok": True, "result": result}
    except Exception as e:
        return {"ok": False, "error": str(e)}
