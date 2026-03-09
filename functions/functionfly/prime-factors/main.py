def handler(event):
    n = event.get("n") if isinstance(event, dict) else None
    if n is None:
        return {"ok": False, "error": "n is required"}
    try:
        n = int(n)
        if n < 2:
            return {"ok": True, "factors": []}
        if n < 0:
            n = -n
        out = []
        d = 2
        while d * d <= n:
            while n % d == 0:
                out.append(d)
                n //= d
            d += 1
        if n > 1:
            out.append(n)
        return {"ok": True, "factors": out}
    except (TypeError, ValueError) as e:
        return {"ok": False, "error": str(e)}
