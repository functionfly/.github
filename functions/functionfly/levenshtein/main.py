def levenshtein_distance(s1: str, s2: str) -> int:
    """Calculate Levenshtein distance between two strings"""
    if len(s1) < len(s2):
        return levenshtein_distance(s2, s1)
    if len(s2) == 0:
        return len(s1)
    prev_row = range(len(s2) + 1)
    for i, c1 in enumerate(s1):
        curr_row = [i + 1]
        for j, c2 in enumerate(s2):
            insertions = prev_row[j + 1] + 1
            deletions = curr_row[j] + 1
            substitutions = prev_row[j] + (c1 != c2)
            curr_row.append(min(insertions, deletions, substitutions))
        prev_row = curr_row
    return prev_row[-1]

def handler(event):
    try:
        string1 = event.get("string1", "") if isinstance(event, dict) else ""
        string2 = event.get("string2", "") if isinstance(event, dict) else ""
        if not string1:
            return {"ok": False, "error": "string1 is required"}
        if not string2:
            return {"ok": False, "error": "string2 is required"}
        distance = levenshtein_distance(string1, string2)
        max_len = max(len(string1), len(string2))
        similarity = 1 - (distance / max_len) if max_len > 0 else 1.0
        return {"ok": True, "distance": distance, "similarity": round(similarity, 2)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
