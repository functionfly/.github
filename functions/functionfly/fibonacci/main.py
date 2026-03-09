def handler(event):
    n = event.get("n") if isinstance(event, dict) else None
    if n is None:
        return {"ok": False, "error": "n is required"}
    try:
        n = int(n)
        if n < 0:
            return {"ok": False, "error": "n must be non-negative"}
        if n <= 1:
            return {"ok": True, "result": n}
        a, b = 0, 1
        for _ in range(2, n + 1):
            a, b = b, a + b
        return {"ok": True, "result": b}
    except (TypeError, ValueError) as e:
        return {"ok": False, "error": str(e)}
