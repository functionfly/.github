def handler(event):
    n = event.get("n") if isinstance(event, dict) else None
    if n is None:
        return {"ok": False, "error": "n is required"}
    try:
        n = int(n)
        if n < 2:
            return {"ok": True, "is_prime": False}
        if n == 2:
            return {"ok": True, "is_prime": True}
        if n % 2 == 0:
            return {"ok": True, "is_prime": False}
        i = 3
        while i * i <= n:
            if n % i == 0:
                return {"ok": True, "is_prime": False}
            i += 2
        return {"ok": True, "is_prime": True}
    except (TypeError, ValueError) as e:
        return {"ok": False, "error": str(e)}
