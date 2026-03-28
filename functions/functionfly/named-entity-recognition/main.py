import re

PERSON_TITLES = {"mr","mrs","ms","dr","prof","sir","dame","rev","capt","maj","col","gen","lt","sgt","ceo","cto","cfo","vp","president","director","manager","officer"}
ORG_SUFFIXES = {"inc","corp","ltd","llc","plc","co","company","group","holdings","ventures","technologies","systems","solutions","services","labs","studio","studios","agency","partners","associates","foundation","institute","university","college","school","hospital","bank","fund","trust"}
PLACE_WORDS = {"city","town","village","country","state","province","district","county","region","street","avenue","road","boulevard","park","lake","river","mountain","ocean","sea","bay","island","valley","forest","desert","peninsula","continent","territory","republic","kingdom","empire"}
MONTHS = {"january","february","march","april","may","june","july","august","september","october","november","december","jan","feb","mar","apr","jun","jul","aug","sep","oct","nov","dec"}


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    if not text or not isinstance(text, str):
        return {"ok": False, "error": "text (string) is required"}
    try:
        t = str(text)
        entities = []

        # Dates
        date_patterns = [
            r'\b(?:January|February|March|April|May|June|July|August|September|October|November|December)\s+\d{1,2},?\s+\d{4}\b',
            r'\b\d{1,2}/\d{1,2}/\d{2,4}\b',
            r'\b\d{4}-\d{2}-\d{2}\b',
            r'\b(?:Jan|Feb|Mar|Apr|Jun|Jul|Aug|Sep|Oct|Nov|Dec)\.?\s+\d{1,2},?\s+\d{4}\b',
        ]
        for pat in date_patterns:
            for m in re.finditer(pat, t, re.IGNORECASE):
                entities.append({"text": m.group(), "label": "DATE", "start": m.start(), "end": m.end()})

        # Money
        for m in re.finditer(r'\$[\d,]+(?:\.\d+)?(?:\s*(?:million|billion|trillion|thousand))?|\b\d+(?:\.\d+)?\s*(?:USD|EUR|GBP|JPY|dollars?|euros?|pounds?)\b', t, re.IGNORECASE):
            entities.append({"text": m.group(), "label": "MONEY", "start": m.start(), "end": m.end()})

        # Percentages
        for m in re.finditer(r'\b\d+(?:\.\d+)?\s*%', t):
            entities.append({"text": m.group(), "label": "PERCENT", "start": m.start(), "end": m.end()})

        # Capitalized phrases (persons, orgs, locations)
        for m in re.finditer(r'\b[A-Z][a-z]+(?:\s+[A-Z][a-z]+)*\b', t):
            phrase = m.group()
            parts = phrase.split()
            first_lower = parts[0].lower()
            last_lower = parts[-1].lower()
            if last_lower in ORG_SUFFIXES:
                label = "ORGANIZATION"
            elif last_lower in PLACE_WORDS or first_lower in PLACE_WORDS:
                label = "LOCATION"
            elif first_lower in PERSON_TITLES or len(parts) >= 2:
                label = "PERSON"
            else:
                label = "MISC"
            entities.append({"text": phrase, "label": label, "start": m.start(), "end": m.end()})

        # Deduplicate by span
        seen = set()
        unique = []
        for e in sorted(entities, key=lambda x: x["start"]):
            key = (e["start"], e["end"])
            if key not in seen:
                seen.add(key)
                unique.append(e)

        by_label = {}
        for e in unique:
            by_label.setdefault(e["label"], []).append(e["text"])

        return {
            "ok": True,
            "result": unique,
            "entities": unique,
            "by_label": by_label,
            "count": len(unique)
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
