import re


PERSON_TITLES = {"mr","mrs","ms","dr","prof","sir","dame","rev","capt","maj","col","gen","lt","sgt"}
ORG_SUFFIXES = {"inc","corp","ltd","llc","plc","co","company","group","holdings","ventures","technologies","systems","solutions","services","labs","studio","studios","agency","partners","associates"}
PLACE_WORDS = {"city","town","village","country","state","province","district","county","region","street","avenue","road","boulevard","park","lake","river","mountain","ocean","sea","bay"}


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    if not text:
        return {"ok": False, "error": "text is required"}
    try:
        t = str(text)
        persons, orgs, locations, other = [], [], [], []
        token_groups = re.findall(r'\b[A-Z][a-z]+(?:\s+[A-Z][a-z]+)*\b', t)
        words_lower = set(t.lower().split())
        for phrase in token_groups:
            parts = phrase.split()
            first_lower = parts[0].lower()
            last_lower = parts[-1].lower()
            if first_lower in PERSON_TITLES or (len(parts) >= 2 and all(p[0].isupper() for p in parts)):
                if last_lower in ORG_SUFFIXES:
                    orgs.append(phrase)
                elif last_lower in PLACE_WORDS or first_lower in PLACE_WORDS:
                    locations.append(phrase)
                elif len(parts) >= 2:
                    persons.append(phrase)
                else:
                    other.append(phrase)
        return {
            "ok": True,
            "result": {"persons": list(set(persons)), "organizations": list(set(orgs)), "locations": list(set(locations))},
            "entities": {
                "PERSON": list(set(persons))[:20],
                "ORGANIZATION": list(set(orgs))[:20],
                "LOCATION": list(set(locations))[:20],
                "OTHER": list(set(other))[:20],
            },
            "note": "Heuristic NER — for production use, integrate spaCy or similar"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
