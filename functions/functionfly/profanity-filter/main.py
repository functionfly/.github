import re


PROFANITY = ["fuck","shit","ass","bitch","cunt","dick","piss","cock","whore","slut","bastard","damn","hell","crap","asshole","motherfucker","bullshit","jackass","douchebag","retard","faggot","nigger","spic","chink","kike","wetback"]


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    replacement = event.get("replacement", "***")
    partial_match = event.get("partial_match", False)
    if not text:
        return {"ok": False, "error": "text is required"}
    try:
        result = str(text)
        found = []
        for word in PROFANITY:
            if partial_match:
                pattern = re.compile(re.escape(word), re.I)
            else:
                pattern = re.compile(r'\b' + re.escape(word) + r'\b', re.I)
            matches = pattern.findall(result)
            if matches:
                found.extend(matches)
                rep = replacement if replacement != "***" else "*" * len(word)
                result = pattern.sub(rep, result)
        has_profanity = len(found) > 0
        return {"ok": True, "result": result, "filtered_text": result, "has_profanity": has_profanity, "count": len(found), "found": list(set(w.lower() for w in found))}
    except Exception as e:
        return {"ok": False, "error": str(e)}
