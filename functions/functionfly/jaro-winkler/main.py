def jaro_similarity(s1: str, s2: str) -> float:
    """Calculate Jaro similarity between two strings"""
    if s1 == s2:
        return 1.0
    len1, len2 = len(s1), len(s2)
    if len1 == 0 or len2 == 0:
        return 0.0
    match_distance = max(len1, len2) // 2 - 1
    s1_matches = [False] * len1
    s2_matches = [False] * len2
    matches = 0
    transpositions = 0
    for i in range(len1):
        start = max(0, i - match_distance)
        end = min(i + match_distance + 1, len2)
        for j in range(start, end):
            if s2_matches[j] or s1[i] != s2[j]:
                continue
            s1_matches[i] = True
            s2_matches[j] = True
            matches += 1
            break
    if matches == 0:
        return 0.0
    k = 0
    for i in range(len1):
        if not s1_matches[i]:
            continue
        while not s2_matches[k]:
            k += 1
        if s1[i] != s2[k]:
            transpositions += 1
        k += 1
    jaro = (matches / len1 + matches / len2 + (matches - transpositions / 2) / matches) / 3
    return jaro

def jaro_winkler(s1: str, s2: str, p: float = 0.1) -> float:
    """Calculate Jaro-Winkler similarity between two strings"""
    jaro = jaro_similarity(s1, s2)
    prefix = 0
    for i in range(min(len(s1), len(s2), 4)):
        if s1[i] == s2[i]:
            prefix += 1
        else:
            break
    return jaro + prefix * p * (1 - jaro)

def handler(event):
    try:
        string1 = event.get("string1", "") if isinstance(event, dict) else ""
        string2 = event.get("string2", "") if isinstance(event, dict) else ""
        if not string1:
            return {"ok": False, "error": "string1 is required"}
        if not string2:
            return {"ok": False, "error": "string2 is required"}
        similarity = jaro_winkler(string1, string2)
        return {"ok": True, "similarity": round(similarity, 2)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
