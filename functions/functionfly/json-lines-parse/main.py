import json


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    skip_empty = event.get("skip_empty", True)
    if not data:
        return {"ok": False, "error": "data is required (JSON Lines string)"}
    try:
        results = []
        errors = []
        for i, line in enumerate(str(data).splitlines()):
            line = line.strip()
            if not line:
                if not skip_empty:
                    errors.append({"line": i + 1, "error": "empty line"})
                continue
            try:
                results.append(json.loads(line))
            except json.JSONDecodeError as e:
                errors.append({"line": i + 1, "error": str(e)})
        return {"ok": True, "result": results, "count": len(results), "errors": errors}
    except Exception as e:
        return {"ok": False, "error": str(e)}
