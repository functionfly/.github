def soundex(s: str) -> str:
    """Generate Soundex code for a string"""
    if not s:
        return ""
    s = s.upper()
    result = s[0]
    mapping = {
        'B': '1', 'F': '1', 'P': '1', 'V': '1',
        'C': '2', 'G': '2', 'J': '2', 'K': '2', 'Q': '2', 'S': '2', 'X': '2', 'Z': '2',
        'D': '3', 'T': '3',
        'L': '4',
        'M': '5', 'N': '5',
        'R': '6'
    }
    for char in s[1:]:
        if char in mapping:
            code = mapping[char]
            if code != result[-1]:
                result += code
    result = result[:4].ljust(4, '0')
    return result

def metaphone(s: str) -> str:
    """Generate Metaphone code for a string"""
    if not s:
        return ""
    s = s.upper()
    result = []
    i = 0
    while i < len(s):
        c = s[i]
        if i > 0 and c == s[i-1]:
            i += 1
            continue
        if c in 'AEIOU':
            if i == 0:
                result.append(c)
        elif c == 'B':
            if not (i > 0 and s[i-1] == 'M'):
                result.append('B')
        elif c == 'C':
            if i + 1 < len(s) and s[i+1] in 'EIY':
                result.append('S')
            else:
                result.append('K')
        elif c == 'D':
            if i + 1 < len(s) and s[i+1] == 'G':
                result.append('J')
                i += 1
            else:
                result.append('T')
        elif c == 'F':
            result.append('F')
        elif c == 'G':
            if i + 1 < len(s) and s[i+1] in 'EIY':
                result.append('J')
            else:
                result.append('K')
        elif c == 'H':
            if i == 0 or s[i-1] not in 'AEIOU':
                result.append('H')
        elif c == 'J':
            result.append('J')
        elif c == 'K':
            if i == 0 or s[i-1] != 'C':
                result.append('K')
        elif c == 'L':
            result.append('L')
        elif c == 'M':
            result.append('M')
        elif c == 'N':
            result.append('N')
        elif c == 'P':
            if i + 1 < len(s) and s[i+1] == 'H':
                result.append('F')
                i += 1
            else:
                result.append('P')
        elif c == 'Q':
            result.append('K')
        elif c == 'R':
            result.append('R')
        elif c == 'S':
            if i + 1 < len(s) and s[i+1] == 'H':
                result.append('X')
                i += 1
            elif i + 2 < len(s) and s[i+1] == 'I' and s[i+2] in 'AO':
                result.append('X')
                i += 2
            else:
                result.append('S')
        elif c == 'T':
            if i + 1 < len(s) and s[i+1] == 'H':
                result.append('0')
                i += 1
            elif i + 2 < len(s) and s[i+1] == 'I' and s[i+2] in 'AO':
                result.append('X')
                i += 2
            else:
                result.append('T')
        elif c == 'V':
            result.append('F')
        elif c == 'W':
            if i == 0 or s[i-1] in 'AEIOU':
                result.append('W')
        elif c == 'X':
            result.append('KS')
        elif c == 'Y':
            if i == 0 or s[i-1] in 'AEIOU':
                result.append('Y')
        elif c == 'Z':
            result.append('S')
        i += 1
    return ''.join(result)

def handler(event):
    try:
        string1 = event.get("string1", "") if isinstance(event, dict) else ""
        string2 = event.get("string2", "") if isinstance(event, dict) else ""
        algorithm = event.get("algorithm", "soundex") if isinstance(event, dict) else "soundex"
        if not string1:
            return {"ok": False, "error": "string1 is required"}
        if not string2:
            return {"ok": False, "error": "string2 is required"}
        if algorithm == "soundex":
            code1 = soundex(string1)
            code2 = soundex(string2)
        elif algorithm == "metaphone":
            code1 = metaphone(string1)
            code2 = metaphone(string2)
        else:
            return {"ok": False, "error": f"unsupported algorithm: {algorithm}"}
        match = code1 == code2
        return {"ok": True, "match": match, "algorithm": algorithm}
    except Exception as e:
        return {"ok": False, "error": str(e)}
