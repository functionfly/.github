import random


def handler(event):
    items = event.get("items") if isinstance(event, dict) else None
    seed = event.get("seed")
    if items is None:
        return {"ok": False, "error": "items is required"}
    if not isinstance(items, (list, tuple)):
        return {"ok": False, "error": "items must be an array"}
    result = list(items)
    if seed is not None:
        try:
            random.seed(int(seed))
        except (TypeError, ValueError):
            pass
    random.shuffle(result)
    return {"ok": True, "result": result}
