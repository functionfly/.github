def hamming_distance(s1: str, s2: str) -> int:
    """Calculate Hamming distance between two strings"""
    if len(s1) != len(s2):
        raise ValueError("Strings must have equal length")
    return sum(c1 != c2 for c1, c2 in zip(s1, s2))

def handler(event):
    try:
        string1 = event.get("string1", "") if isinstance(event, dict) else ""
        string2 = event.get("string2", "") if isinstance(event, dict) else ""
        if not string1:
            return {"ok": False, "error": "string1 is required"}
        if not string2:
            return {"ok": False, "error": "string2 is required"}
        if len(string1) != len(string2):
            return {"ok": False, "error": "strings must have equal length"}
        distance = hamming_distance(string1, string2)
        return {"ok": True, "distance": distance}
    except Exception as e:
        return {"ok": False, "error": str(e)}
