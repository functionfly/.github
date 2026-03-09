import math

def handler(event):
    n = event.get("n") if isinstance(event, dict) else None
    if n is None:
        return {"ok": False, "error": "n is required"}
    try:
        n = int(n)
        if n < 0:
            return {"ok": False, "error": "n must be non-negative"}
        return {"ok": True, "result": math.factorial(n)}
    except (TypeError, ValueError) as e:
        return {"ok": False, "error": str(e)}
