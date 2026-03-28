def soundex(s: str) -> str:
    """Generate Soundex code for a string"""
    if not s:
        return ""
    s = s.upper()
    # Keep first letter
    result = s[0]
    # Mapping table
    mapping = {
        'B': '1', 'F': '1', 'P': '1', 'V': '1',
        'C': '2', 'G': '2', 'J': '2', 'K': '2', 'Q': '2', 'S': '2', 'X': '2', 'Z': '2',
        'D': '3', 'T': '3',
        'L': '4',
        'M': '5', 'N': '5',
        'R': '6'
    }
    # Process remaining letters
    for char in s[1:]:
        if char in mapping:
            code = mapping[char]
            if code != result[-1]:
                result += code
    # Pad with zeros or truncate to 4 characters
    result = result[:4].ljust(4, '0')
    return result

def handler(event):
    try:
        string = event.get("string", "") if isinstance(event, dict) else ""
        if not string:
            return {"ok": False, "error": "string is required"}
        result = soundex(string)
        return {"ok": True, "soundex": result}
    except Exception as e:
        return {"ok": False, "error": str(e)}
