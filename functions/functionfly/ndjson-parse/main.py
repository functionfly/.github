import json


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    if not data:
        return {"ok": False, "error": "data is required (NDJSON string)"}
    try:
        results, errors = [], []
        for i, line in enumerate(str(data).splitlines()):
            line = line.strip()
            if not line:
                continue
            try:
                results.append(json.loads(line))
            except json.JSONDecodeError as e:
                errors.append({"line": i+1, "error": str(e)})
        return {"ok": True, "result": results, "count": len(results), "errors": errors}
    except Exception as e:
        return {"ok": False, "error": str(e)}
