from urllib.parse import parse_qs, urlparse


def handler(event):
    if isinstance(event, dict):
        query = event.get("query", event.get("q", ""))
    else:
        query = str(event) if event is not None else ""

    if query is None:
        return {"ok": False, "error": "Input 'query' is required"}

    q = str(query).strip()
    if q.startswith("?"):
        q = q[1:]
    if "?" in q:
        q = q.split("?", 1)[1]

    parsed = parse_qs(q, keep_blank_values=True)
    params = {}
    for k, v in parsed.items():
        if len(v) == 1:
            params[k] = v[0]
        else:
            params[k] = v
    return {"ok": True, "params": params}

