import random


def handler(event):
    items = event.get("items") if isinstance(event, dict) else None
    n = event.get("n")
    seed = event.get("seed")
    if items is None:
        return {"ok": False, "error": "items is required"}
    if not isinstance(items, (list, tuple)):
        return {"ok": False, "error": "items must be an array"}
    if n is None:
        return {"ok": False, "error": "n is required"}
    try:
        n = int(n)
    except (TypeError, ValueError):
        return {"ok": False, "error": "n must be an integer"}
    if n < 1:
        return {"ok": False, "error": "n must be at least 1"}
    if seed is not None:
        try:
            random.seed(int(seed))
        except (TypeError, ValueError):
            pass
    arr = list(items)
    if n > len(arr):
        return {"ok": False, "error": "n cannot exceed array length"}
    result = random.sample(arr, n)
    return {"ok": True, "result": result}
