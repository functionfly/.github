import statistics

def handler(event):
    vals = event.get("values") if isinstance(event, dict) else None
    sample = event.get("sample", False) if isinstance(event, dict) else False
    if vals is None:
        return {"ok": False, "error": "values is required"}
    if not isinstance(vals, (list, tuple)):
        return {"ok": False, "error": "values must be an array"}
    try:
        nums = [float(x) for x in vals]
        if len(nums) < 2:
            return {"ok": False, "error": "need at least 2 values"}
        var = statistics.variance(nums) if sample else statistics.pvariance(nums)
        return {"ok": True, "variance": var}
    except (TypeError, ValueError) as e:
        return {"ok": False, "error": str(e)}
