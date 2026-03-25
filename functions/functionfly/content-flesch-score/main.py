import re


def _count_syllables(word):
    word = word.lower().rstrip('.,!?;:')
    vowels = re.findall(r'[aeiou]', word)
    diphthongs = len(re.findall(r'[aeiou]{2}', word))
    n = max(1, len(vowels) - diphthongs)
    if word.endswith('e') and len(word) > 2 and word[-2] not in 'aeiou':
        n = max(1, n - 1)
    return n


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    if not text:
        return {"ok": False, "error": "text is required"}
    try:
        stripped = re.sub(r'<[^>]+>', ' ', str(text))
        sentences = max(1, len(re.findall(r'[.!?]+', stripped)))
        words = re.findall(r'\b[a-zA-Z]+\b', stripped)
        if not words:
            return {"ok": False, "error": "No words found in text"}
        total_syllables = sum(_count_syllables(w) for w in words)
        asl = len(words) / sentences
        asw = total_syllables / len(words)
        fre = round(206.835 - 1.015 * asl - 84.6 * asw, 2)
        fre = max(0, min(121, fre))
        if fre >= 90: level = "5th grade (Very Easy)"
        elif fre >= 80: level = "6th grade (Easy)"
        elif fre >= 70: level = "7th grade (Fairly Easy)"
        elif fre >= 60: level = "8th-9th grade (Standard)"
        elif fre >= 50: level = "10th-12th grade (Fairly Difficult)"
        elif fre >= 30: level = "College (Difficult)"
        else: level = "College Graduate (Very Difficult)"
        fkgl = round(0.39 * asl + 11.8 * asw - 15.59, 1)
        return {
            "ok": True,
            "result": fre,
            "flesch_reading_ease": fre,
            "flesch_kincaid_grade": fkgl,
            "level": level,
            "avg_sentence_length": round(asl, 2),
            "avg_syllables_per_word": round(asw, 2),
            "word_count": len(words),
            "sentence_count": sentences
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
