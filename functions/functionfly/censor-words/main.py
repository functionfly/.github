import re


# Minimal default blocklist (common profanity); can be overridden by input
_DEFAULT_WORDS = {"damn", "hell", "crap", "stupid", "idiot", "dumb", "bad"}


def handler(event):
    if isinstance(event, dict):
        text = event.get("text", event.get("data", ""))
        replacement = event.get("replacement", "***")
        words = event.get("words")
    else:
        text = str(event) if event is not None else ""
        replacement = "***"
        words = None

    if text is None:
        return {"ok": False, "error": "Input 'text' is required"}

    blocklist = set(w.strip().lower() for w in words) if isinstance(words, list) and words else _DEFAULT_WORDS
    blocklist = {w for w in blocklist if w}
    if not blocklist:
        return {"ok": True, "result": str(text), "replacements_count": 0}

    rep = str(replacement) if replacement is not None else "***"
    pattern = re.compile(r"\b(" + "|".join(re.escape(w) for w in blocklist) + r")\b", re.IGNORECASE)
    result, n = pattern.subn(lambda m: rep, str(text))
    return {"ok": True, "result": result, "replacements_count": n}

