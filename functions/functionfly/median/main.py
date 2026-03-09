import statistics

def handler(event):
    vals = event.get("values") if isinstance(event, dict) else None
    if vals is None:
        return {"ok": False, "error": "values is required"}
    if not isinstance(vals, (list, tuple)):
        return {"ok": False, "error": "values must be an array"}
    try:
        nums = [float(x) for x in vals]
        if not nums:
            return {"ok": False, "error": "values must not be empty"}
        return {"ok": True, "median": statistics.median(nums)}
    except (TypeError, ValueError) as e:
        return {"ok": False, "error": str(e)}
