from urllib.parse import urlencode


def handler(event):
    if isinstance(event, dict):
        params = event.get("params", event.get("data", {}))
        prefix_question_mark = event.get("prefix_question_mark", True)
    else:
        params = {}
        prefix_question_mark = True

    if params is None:
        return {"ok": False, "error": "Input 'params' is required"}

    if not isinstance(params, dict):
        return {"ok": False, "error": "Input 'params' must be an object"}

    flat = []
    for k, v in params.items():
        if isinstance(v, list):
            for item in v:
                flat.append((k, str(item)))
        else:
            flat.append((k, str(v) if v is not None else ""))

    query = urlencode(flat)
    if prefix_question_mark and query:
        query = "?" + query
    return {"ok": True, "query": query}

