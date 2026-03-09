import re


def _pluralize(word):
    w = word.strip()
    if not w:
        return w
    lower = w.lower()
    irregular = {
        "child": "children", "person": "people", "man": "men", "woman": "women",
        "tooth": "teeth", "foot": "feet", "mouse": "mice", "goose": "geese",
        "ox": "oxen", "leaf": "leaves", "life": "lives", "knife": "knives",
        "half": "halves", "self": "selves", "sheep": "sheep", "deer": "deer",
    }
    if lower in irregular:
        return w[:-len(lower)] + irregular[lower] if w != w.lower() else irregular[lower]
    if w.endswith(("s", "x", "z")) or w.endswith(("ch", "sh")):
        return w + "es"
    if w.endswith("y") and len(w) > 1 and w[-2] not in "aeiou":
        return w[:-1] + "ies"
    if w.endswith("f"):
        return w[:-1] + "ves"
    if w.endswith("fe"):
        return w[:-2] + "ves"
    if w.endswith("o") and len(w) > 1 and w[-2] not in "aeiou":
        return w + "es"
    return w + "s"


def handler(event):
    if isinstance(event, dict):
        word = event.get("word", event.get("text", ""))
        count = event.get("count", 2)
    else:
        word = str(event) if event is not None else ""
        count = 2

    if word is None:
        return {"ok": False, "error": "Input 'word' is required"}

    try:
        count = int(count)
    except (TypeError, ValueError):
        count = 2

    if count == 1:
        return {"ok": True, "result": str(word).strip()}
    return {"ok": True, "result": _pluralize(str(word))}

