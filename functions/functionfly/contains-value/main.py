def handler(event):
    value = event.get("value") if isinstance(event, dict) else None
    search = event.get("search")
    case_sensitive = event.get("case_sensitive", True)

    if value is None and "value" not in event:
        return {"ok": False, "error": "value is required"}
    if search is None:
        return {"ok": False, "error": "search is required"}

    if isinstance(value, (list, tuple)):
        result = search in value
    elif isinstance(value, dict):
        result = search in value.values()
    elif isinstance(value, str):
        if case_sensitive:
            result = str(search) in value
        else:
            result = str(search).lower() in value.lower()
    else:
        result = str(search) in str(value)

    return {"ok": True, "value": value, "result": result, "search": search}
