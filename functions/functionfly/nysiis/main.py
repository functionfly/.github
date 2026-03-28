def nysiis(s: str) -> str:
    """Generate NYSIIS code for a string"""
    if not s:
        return ""
    s = s.upper()
    # Remove non-alpha characters
    s = ''.join(c for c in s if c.isalpha())
    if not s:
        return ""
    # Initial translations
    if s.startswith('MAC'):
        s = 'MCC' + s[3:]
    elif s.startswith('KN'):
        s = 'NN' + s[2:]
    elif s.startswith('K'):
        s = 'C' + s[1:]
    elif s.startswith('PH'):
        s = 'FF' + s[2:]
    elif s.startswith('PF'):
        s = 'FF' + s[2:]
    elif s.startswith('SCH'):
        s = 'SSS' + s[3:]
    # Final translations
    if s.endswith('EE'):
        s = s[:-2] + 'Y'
    elif s.endswith('IE'):
        s = s[:-2] + 'Y'
    elif s.endswith('DT'):
        s = s[:-2] + 'D'
    elif s.endswith('RT'):
        s = s[:-2] + 'D'
    elif s.endswith('RD'):
        s = s[:-2] + 'D'
    elif s.endswith('NT'):
        s = s[:-2] + 'D'
    elif s.endswith('ND'):
        s = s[:-2] + 'D'
    # Translate first character
    first_char = s[0]
    # Translate remaining characters
    result = [first_char]
    i = 1
    while i < len(s):
        c = s[i]
        # Translate vowels to A
        if c in 'AEIOU':
            result.append('A')
        # Translate consonants
        elif c == 'Q':
            result.append('G')
        elif c == 'Z':
            result.append('S')
        elif c == 'M':
            result.append('N')
        elif c == 'K':
            if i + 1 < len(s) and s[i+1] == 'N':
                result.append('N')
            else:
                result.append('C')
        elif c == 'S' and i + 1 < len(s) and s[i+1] == 'C':
            result.append('S')
            i += 1
        elif c == 'P' and i + 1 < len(s) and s[i+1] == 'H':
            result.append('F')
            i += 1
        elif c == 'H' and (i == 0 or s[i-1] not in 'AEIOU'):
            result.append(result[-1] if result else '')
        elif c == 'W' and (i == 0 or s[i-1] in 'AEIOU'):
            result.append(result[-1] if result else '')
        else:
            result.append(c)
        i += 1
    # Remove consecutive duplicates
    final_result = [result[0]]
    for c in result[1:]:
        if c != final_result[-1]:
            final_result.append(c)
    # Remove trailing A
    if final_result and final_result[-1] == 'A':
        final_result.pop()
    return ''.join(final_result)

def handler(event):
    try:
        string = event.get("string", "") if isinstance(event, dict) else ""
        if not string:
            return {"ok": False, "error": "string is required"}
        result = nysiis(string)
        return {"ok": True, "nysiis": result}
    except Exception as e:
        return {"ok": False, "error": str(e)}
