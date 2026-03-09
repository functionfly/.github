def handler(event):
    if isinstance(event, dict):
        a = event.get("a", "")
        b = event.get("b", "")
    else:
        a, b = "", ""

    if a is None or b is None:
        return {"ok": False, "error": "Inputs 'a' and 'b' are required"}

    s, t = str(a), str(b)
    n, m = len(s), len(t)
    if n == 0:
        return {"ok": True, "distance": m}
    if m == 0:
        return {"ok": True, "distance": n}

    prev = list(range(m + 1))
    for i in range(1, n + 1):
        curr = [i]
        for j in range(1, m + 1):
            cost = 0 if s[i - 1] == t[j - 1] else 1
            curr.append(min(prev[j] + 1, curr[j - 1] + 1, prev[j - 1] + cost))
        prev = curr
    return {"ok": True, "distance": prev[m]}

