import re

# Rule-based lemmatization dictionary
LEMMA_DICT = {
    # Irregular verbs
    "am": "be", "is": "be", "are": "be", "was": "be", "were": "be", "been": "be", "being": "be",
    "has": "have", "had": "have", "having": "have",
    "does": "do", "did": "do", "done": "do", "doing": "do",
    "went": "go", "gone": "go", "going": "go", "goes": "go",
    "saw": "see", "seen": "see", "seeing": "see", "sees": "see",
    "said": "say", "says": "say", "saying": "say",
    "got": "get", "gotten": "get", "getting": "get", "gets": "get",
    "made": "make", "makes": "make", "making": "make",
    "came": "come", "comes": "come", "coming": "come",
    "took": "take", "taken": "take", "takes": "take", "taking": "take",
    "knew": "know", "known": "know", "knows": "know", "knowing": "know",
    "thought": "think", "thinks": "think", "thinking": "think",
    "found": "find", "finds": "find", "finding": "find",
    "gave": "give", "given": "give", "gives": "give", "giving": "give",
    "told": "tell", "tells": "tell", "telling": "tell",
    "became": "become", "becomes": "become", "becoming": "become",
    "showed": "show", "shown": "show", "shows": "show", "showing": "show",
    "left": "leave", "leaves": "leave", "leaving": "leave",
    "felt": "feel", "feels": "feel", "feeling": "feel",
    "put": "put", "puts": "put", "putting": "put",
    "brought": "bring", "brings": "bring", "bringing": "bring",
    "began": "begin", "begun": "begin", "begins": "begin", "beginning": "begin",
    "kept": "keep", "keeps": "keep", "keeping": "keep",
    "held": "hold", "holds": "hold", "holding": "hold",
    "wrote": "write", "written": "write", "writes": "write", "writing": "write",
    "stood": "stand", "stands": "stand", "standing": "stand",
    "heard": "hear", "hears": "hear", "hearing": "hear",
    "let": "let", "lets": "let", "letting": "let",
    "meant": "mean", "means": "mean", "meaning": "mean",
    "set": "set", "sets": "set", "setting": "set",
    "met": "meet", "meets": "meet", "meeting": "meet",
    "ran": "run", "runs": "run", "running": "run",
    "paid": "pay", "pays": "pay", "paying": "pay",
    "sat": "sit", "sits": "sit", "sitting": "sit",
    "spoke": "speak", "spoken": "speak", "speaks": "speak", "speaking": "speak",
    "lay": "lie", "lain": "lie", "lies": "lie", "lying": "lie",
    "led": "lead", "leads": "lead", "leading": "lead",
    "read": "read", "reads": "read", "reading": "read",
    "grew": "grow", "grown": "grow", "grows": "grow", "growing": "grow",
    "lost": "lose", "loses": "lose", "losing": "lose",
    "fell": "fall", "fallen": "fall", "falls": "fall", "falling": "fall",
    "sent": "send", "sends": "send", "sending": "send",
    "built": "build", "builds": "build", "building": "build",
    "spent": "spend", "spends": "spend", "spending": "spend",
    "cut": "cut", "cuts": "cut", "cutting": "cut",
    "hit": "hit", "hits": "hit", "hitting": "hit",
    "drove": "drive", "driven": "drive", "drives": "drive", "driving": "drive",
    "bought": "buy", "buys": "buy", "buying": "buy",
    "wore": "wear", "worn": "wear", "wears": "wear", "wearing": "wear",
    "chose": "choose", "chosen": "choose", "chooses": "choose", "choosing": "choose",
    "broke": "break", "broken": "break", "breaks": "break", "breaking": "break",
    # Irregular nouns
    "mice": "mouse", "geese": "goose", "feet": "foot", "teeth": "tooth",
    "men": "man", "women": "woman", "children": "child", "people": "person",
    "oxen": "ox", "cacti": "cactus", "fungi": "fungus", "alumni": "alumnus",
    "criteria": "criterion", "phenomena": "phenomenon", "data": "datum",
    "analyses": "analysis", "bases": "basis", "crises": "crisis",
    "diagnoses": "diagnosis", "hypotheses": "hypothesis", "parentheses": "parenthesis",
    "theses": "thesis", "matrices": "matrix", "indices": "index",
    "appendices": "appendix", "vertices": "vertex", "axes": "axis",
    # Comparative/superlative adjectives
    "better": "good", "best": "good", "worse": "bad", "worst": "bad",
    "more": "many", "most": "many", "less": "little", "least": "little",
    "further": "far", "furthest": "far", "farther": "far", "farthest": "far",
    "older": "old", "oldest": "old", "elder": "old", "eldest": "old",
}


def _apply_rules(word):
    w = word.lower()
    if w in LEMMA_DICT:
        return LEMMA_DICT[w]
    # Verb rules
    if w.endswith("ies") and len(w) > 4:
        return w[:-3] + "y"
    if w.endswith("ied") and len(w) > 4:
        return w[:-3] + "y"
    if w.endswith("ying") and len(w) > 5:
        return w[:-4] + "y"
    if w.endswith("ing") and len(w) > 5:
        stem = w[:-3]
        if stem.endswith(stem[-1]) and stem[-1] not in "aeiou" and len(stem) > 3:
            return stem[:-1]
        if stem.endswith("e"):
            return stem
        return stem + "e" if len(stem) < 3 else stem
    if w.endswith("ed") and len(w) > 4:
        stem = w[:-2]
        if stem.endswith(stem[-1]) and stem[-1] not in "aeiou" and len(stem) > 3:
            return stem[:-1]
        if stem.endswith("e"):
            return stem
        return stem
    # Noun plural rules
    if w.endswith("ves") and len(w) > 4:
        return w[:-3] + "f"
    if w.endswith("ies") and len(w) > 4:
        return w[:-3] + "y"
    if w.endswith("ses") or w.endswith("xes") or w.endswith("zes") or w.endswith("ches") or w.endswith("shes"):
        return w[:-2]
    if w.endswith("s") and not w.endswith("ss") and len(w) > 3:
        return w[:-1]
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
            lemma = _apply_rules(word)
            results.append({"word": word, "lemma": lemma})
        return {
            "ok": True,
            "result": results,
            "lemmas": results,
            "count": len(results)
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
