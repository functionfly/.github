def _singularize(word):
    w = word.strip()
    if not w:
        return w
    irregular = {
        "children": "child", "people": "person", "men": "man", "women": "woman",
        "teeth": "tooth", "feet": "foot", "mice": "mouse", "geese": "goose",
        "oxen": "ox", "leaves": "leaf", "lives": "life", "knives": "knife",
        "halves": "half", "selves": "self", "sheep": "sheep", "deer": "deer",
    }
    lower = w.lower()
    if lower in irregular:
        return w[:-len(lower)] + irregular[lower] if w != w.lower() else irregular[lower]
    if lower in ("sheep", "deer"):
        return w
    if w.endswith("ies") and len(w) > 3 and w[-4] not in "aeiou":
        return w[:-3] + "y"
    if w.endswith("ves"):
        return w[:-3] + "f"
    if w.endswith("es") and len(w) > 2 and w[-3] in "sxzch":
        return w[:-2]
    if w.endswith("s") and len(w) > 1 and w[-2] != "s":
        return w[:-1]
    return w


def handler(event):
    if isinstance(event, dict):
        word = event.get("word", event.get("text", ""))
    else:
        word = str(event) if event is not None else ""

    if word is None:
        return {"ok": False, "error": "Input 'word' is required"}

    return {"ok": True, "result": _singularize(str(word))}

