from collections import Counter


def handler(event):
    items = event.get("items") if isinstance(event, dict) else None
    if items is None:
        return {"ok": False, "error": "items is required"}
    if not isinstance(items, (list, tuple)):
        return {"ok": False, "error": "items must be an array"}
    try:
        counts = Counter(items)
        duplicates = [x for x, c in counts.items() if c > 1]
    except TypeError:
        seen = {}
        dupes = []
        for x in items:
            k = id(x)
            if k in seen:
                if seen[k] == 1:
                    dupes.append(x)
                seen[k] += 1
            else:
                seen[k] = 1
        duplicates = dupes
    return {"ok": True, "duplicates": list(duplicates)}
