import re

# Porter Stemmer rules (simplified)
def _porter_stem(word):
    w = word.lower()
    if len(w) <= 2:
        return w
    # Step 1a
    if w.endswith("sses"):
        w = w[:-2]
    elif w.endswith("ies"):
        w = w[:-2]
    elif w.endswith("ss"):
        pass
    elif w.endswith("s") and len(w) > 3:
        w = w[:-1]
    # Step 1b
    if w.endswith("eed"):
        if len(w) > 4:
            w = w[:-1]
    elif w.endswith("ed"):
        stem = w[:-2]
        if re.search(r'[aeiou]', stem):
            w = stem
            if w.endswith("at") or w.endswith("bl") or w.endswith("iz"):
                w += "e"
            elif re.search(r'([^aeiou])\1$', w) and not w.endswith(("l","s","z")):
                w = w[:-1]
    elif w.endswith("ing"):
        stem = w[:-3]
        if re.search(r'[aeiou]', stem):
            w = stem
            if w.endswith("at") or w.endswith("bl") or w.endswith("iz"):
                w += "e"
            elif re.search(r'([^aeiou])\1$', w) and not w.endswith(("l","s","z")):
                w = w[:-1]
    # Step 1c
    if w.endswith("y") and re.search(r'[aeiou]', w[:-1]):
        w = w[:-1] + "i"
    # Step 2
    step2 = [("ational","ate"),("tional","tion"),("enci","ence"),("anci","ance"),
             ("izer","ize"),("abli","able"),("alli","al"),("entli","ent"),
             ("eli","e"),("ousli","ous"),("ization","ize"),("ation","ate"),
             ("ator","ate"),("alism","al"),("iveness","ive"),("fulness","ful"),
             ("ousness","ous"),("aliti","al"),("iviti","ive"),("biliti","ble")]
    for suffix, replacement in step2:
        if w.endswith(suffix) and len(w) > len(suffix) + 1:
            w = w[:-len(suffix)] + replacement
            break
    # Step 3
    step3 = [("icate","ic"),("ative",""),("alize","al"),("iciti","ic"),
             ("ical","ic"),("ful",""),("ness","")]
    for suffix, replacement in step3:
        if w.endswith(suffix) and len(w) > len(suffix) + 1:
            w = w[:-len(suffix)] + replacement
            break
    # Step 4
    step4 = ["al","ance","ence","er","ic","able","ible","ant","ement","ment",
             "ent","ion","ou","ism","ate","iti","ous","ive","ize"]
    for suffix in step4:
        if w.endswith(suffix) and len(w) > len(suffix) + 2:
            if suffix == "ion" and w[-4] in "st":
                w = w[:-len(suffix)]
            elif suffix != "ion":
                w = w[:-len(suffix)]
            break
    # Step 5a
    if w.endswith("e"):
        if len(w) > 4:
            w = w[:-1]
    # Step 5b
    if re.search(r'([^aeiou])\1$', w) and w.endswith("l") and len(w) > 3:
        w = w[:-1]
    return w


def handler(event):
    words = event.get("words") if isinstance(event, dict) else None
    if not words or not isinstance(words, list):
        return {"ok": False, "error": "words (array of strings) is required"}
    try:
        results = []
        for word in words:
            if not isinstance(word, str):
                continue
            stem = _porter_stem(word)
            results.append({"word": word, "stem": stem})
        return {
            "ok": True,
            "result": results,
            "stems": results,
            "count": len(results)
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
