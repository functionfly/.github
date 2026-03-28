import re

SUBJECTS = {"i","you","he","she","it","we","they","who","what","which","that","this","these","those"}
OBJECTS = {"me","him","her","us","them","whom","what","which","that","this","these","those"}
PREPOSITIONS = {"in","on","at","to","for","of","with","by","from","as","into","through","during","before","after","above","below","between","out","off","over","under","about","against","along","among","around","behind","beside","besides","beyond","despite","down","except","inside","near","outside","past","since","throughout","toward","towards","under","underneath","unlike","until","up","upon","within","without"}
DETERMINERS = {"the","a","an","this","that","these","those","my","your","his","her","its","our","their","some","any","each","every","all","both","few","many","much","several","no","another","other"}
AUXILIARIES = {"am","is","are","was","were","be","been","being","have","has","had","do","does","did","will","would","could","should","may","might","shall","can","must"}


def _simple_parse(words):
    """Simple rule-based dependency parsing."""
    deps = []
    root_idx = -1
    # Find main verb (root)
    for i, word in enumerate(words):
        w = word.lower()
        if w in AUXILIARIES:
            continue
        if re.match(r'\b(ing|ed|s|es)\b', w[-3:] if len(w) > 3 else w):
            root_idx = i
            break
    if root_idx == -1:
        # Find any verb-like word
        for i, word in enumerate(words):
            if len(word) > 3 and not word.lower() in DETERMINERS and not word.lower() in PREPOSITIONS:
                root_idx = i
                break
    if root_idx == -1:
        root_idx = 0

    for i, word in enumerate(words):
        w = word.lower()
        if i == root_idx:
            deps.append({"word": word, "dep": "ROOT", "head": word, "head_idx": i})
        elif w in DETERMINERS:
            # Find the noun it modifies (next noun-like word)
            head_idx = min(i + 1, len(words) - 1)
            deps.append({"word": word, "dep": "det", "head": words[head_idx], "head_idx": head_idx})
        elif w in PREPOSITIONS:
            deps.append({"word": word, "dep": "prep", "head": words[root_idx], "head_idx": root_idx})
        elif w in AUXILIARIES:
            deps.append({"word": word, "dep": "aux", "head": words[root_idx], "head_idx": root_idx})
        elif w in SUBJECTS and i < root_idx:
            deps.append({"word": word, "dep": "nsubj", "head": words[root_idx], "head_idx": root_idx})
        elif w in OBJECTS and i > root_idx:
            deps.append({"word": word, "dep": "dobj", "head": words[root_idx], "head_idx": root_idx})
        elif i < root_idx:
            # Before root: likely subject or modifier
            if i > 0 and words[i-1].lower() in DETERMINERS:
                deps.append({"word": word, "dep": "nsubj", "head": words[root_idx], "head_idx": root_idx})
            else:
                deps.append({"word": word, "dep": "amod", "head": words[min(i+1, len(words)-1)], "head_idx": min(i+1, len(words)-1)})
        else:
            # After root: likely object or complement
            if i > 0 and words[i-1].lower() in PREPOSITIONS:
                deps.append({"word": word, "dep": "pobj", "head": words[i-1], "head_idx": i-1})
            else:
                deps.append({"word": word, "dep": "dobj", "head": words[root_idx], "head_idx": root_idx})
    return deps


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    if not text or not isinstance(text, str):
        return {"ok": False, "error": "text (string) is required"}
    try:
        # Process first sentence only
        sentences = re.split(r'(?<=[.!?])\s+', text.strip())
        sentence = sentences[0] if sentences else text
        words = re.findall(r'\b\w+\b', sentence)
        if not words:
            return {"ok": False, "error": "No words found in text"}
        deps = _simple_parse(words)
        return {
            "ok": True,
            "result": deps,
            "dependencies": deps,
            "sentence": sentence,
            "note": "Rule-based dependency parsing — for production use, integrate spaCy or similar"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
