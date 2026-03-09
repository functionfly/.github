def handler(event):
    start = event.get("start", 0)
    stop = event.get("stop")
    step = event.get("step", 1)

    if stop is None:
        return {"ok": False, "error": "stop is required"}

    if not isinstance(start, (int, float)):
        return {"ok": False, "error": "start must be a number"}

    if not isinstance(stop, (int, float)):
        return {"ok": False, "error": "stop must be a number"}

    if not isinstance(step, (int, float)) or step == 0:
        return {"ok": False, "error": "step must be a non-zero number"}

    try:
        result = []
        current = start

        if step > 0:
            while current < stop:
                result.append(current)
                current += step
        else:
            while current > stop:
                result.append(current)
                current += step

        return {
            "ok": True,
            "result": result,
            "count": len(result),
            "start": start,
            "stop": stop,
            "step": step
        }

    except Exception as e:
        return {"ok": False, "error": f"failed to create range: {str(e)}"}