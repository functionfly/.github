def _levenshtein(s, t):
    n, m = len(s), len(t)
    if n == 0:
        return m
    if m == 0:
        return n
    prev = list(range(m + 1))
    for i in range(1, n + 1):
        curr = [i]
        for j in range(1, m + 1):
            cost = 0 if s[i - 1] == t[j - 1] else 1
            curr.append(min(prev[j] + 1, curr[j - 1] + 1, prev[j - 1] + cost))
        prev = curr
    return prev[m]


def handler(event):
    if isinstance(event, dict):
        a = event.get("a", "")
        b = event.get("b", "")
    else:
        a, b = "", ""

    if a is None or b is None:
        return {"ok": False, "error": "Inputs 'a' and 'b' are required"}

    s, t = str(a), str(b)
    if len(s) == 0 and len(t) == 0:
        return {"ok": True, "similarity": 1.0}
    dist = _levenshtein(s, t)
    max_len = max(len(s), len(t))
    similarity = 1.0 - (dist / max_len)
    return {"ok": True, "similarity": round(similarity, 4)}

