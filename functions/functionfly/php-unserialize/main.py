def _unserialize(s, pos=0):
    if pos >= len(s):
        raise ValueError("unexpected end of input")
    t = s[pos]
    if t == "N":
        return None, pos + 2
    if t == "b":
        return bool(int(s[pos+2])), pos + 4
    if t == "i":
        end = s.index(";", pos+2)
        return int(s[pos+2:end]), end + 1
    if t == "d":
        end = s.index(";", pos+2)
        return float(s[pos+2:end]), end + 1
    if t == "s":
        colon1 = s.index(":", pos+2)
        length = int(s[pos+2:colon1])
        start = colon1 + 2
        val = s[start:start+length]
        return val, start + length + 2
    if t == "a":
        colon1 = s.index(":", pos+2)
        count = int(s[pos+2:colon1])
        pos2 = colon1 + 2
        result = {}
        for _ in range(count):
            key, pos2 = _unserialize(s, pos2)
            val, pos2 = _unserialize(s, pos2)
            result[key] = val
        if all(isinstance(k, int) for k in result.keys()):
            n = len(result)
            if sorted(result.keys()) == list(range(n)):
                result = [result[i] for i in range(n)]
        return result, pos2 + 1
    raise ValueError(f"unknown type indicator: {t!r} at pos {pos}")


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    if not data:
        return {"ok": False, "error": "data is required (PHP serialized string)"}
    try:
        result, _ = _unserialize(str(data))
        return {"ok": True, "result": result}
    except Exception as e:
        return {"ok": False, "error": str(e)}
