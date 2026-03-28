def metaphone(s: str) -> str:
    """Generate Metaphone code for a string"""
    if not s:
        return ""
    s = s.upper()
    result = []
    i = 0
    while i < len(s):
        c = s[i]
        # Skip duplicate consecutive letters
        if i > 0 and c == s[i-1]:
            i += 1
            continue
        # Vowels only at start
        if c in 'AEIOU':
            if i == 0:
                result.append(c)
        # B
        elif c == 'B':
            if not (i > 0 and s[i-1] == 'M'):
                result.append('B')
        # C
        elif c == 'C':
            if i + 1 < len(s) and s[i+1] in 'EIY':
                result.append('S')
            else:
                result.append('K')
        # D
        elif c == 'D':
            if i + 1 < len(s) and s[i+1] == 'G':
                result.append('J')
                i += 1
            else:
                result.append('T')
        # F
        elif c == 'F':
            result.append('F')
        # G
        elif c == 'G':
            if i + 1 < len(s) and s[i+1] in 'EIY':
                result.append('J')
            else:
                result.append('K')
        # H
        elif c == 'H':
            if i == 0 or s[i-1] not in 'AEIOU':
                result.append('H')
        # J
        elif c == 'J':
            result.append('J')
        # K
        elif c == 'K':
            if i == 0 or s[i-1] != 'C':
                result.append('K')
        # L
        elif c == 'L':
            result.append('L')
        # M
        elif c == 'M':
            result.append('M')
        # N
        elif c == 'N':
            result.append('N')
        # P
        elif c == 'P':
            if i + 1 < len(s) and s[i+1] == 'H':
                result.append('F')
                i += 1
            else:
                result.append('P')
        # Q
        elif c == 'Q':
            result.append('K')
        # R
        elif c == 'R':
            result.append('R')
        # S
        elif c == 'S':
            if i + 1 < len(s) and s[i+1] == 'H':
                result.append('X')
                i += 1
            elif i + 2 < len(s) and s[i+1] == 'I' and s[i+2] in 'AO':
                result.append('X')
                i += 2
            else:
                result.append('S')
        # T
        elif c == 'T':
            if i + 1 < len(s) and s[i+1] == 'H':
                result.append('0')
                i += 1
            elif i + 2 < len(s) and s[i+1] == 'I' and s[i+2] in 'AO':
                result.append('X')
                i += 2
            else:
                result.append('T')
        # V
        elif c == 'V':
            result.append('F')
        # W
        elif c == 'W':
            if i == 0 or s[i-1] in 'AEIOU':
                result.append('W')
        # X
        elif c == 'X':
            result.append('KS')
        # Y
        elif c == 'Y':
            if i == 0 or s[i-1] in 'AEIOU':
                result.append('Y')
        # Z
        elif c == 'Z':
            result.append('S')
        i += 1
    return ''.join(result)

def handler(event):
    try:
        string = event.get("string", "") if isinstance(event, dict) else ""
        if not string:
            return {"ok": False, "error": "string is required"}
        result = metaphone(string)
        return {"ok": True, "metaphone": result}
    except Exception as e:
        return {"ok": False, "error": str(e)}
