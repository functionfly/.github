import re
import codecs


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None

    if data is None:
        return {"ok": False, "error": "data is required"}
    try:
        s = str(data)
        result = re.sub(
            r'\\[Uu]([0-9a-fA-F]{4,8})',
            lambda m: chr(int(m.group(1), 16)),
            s
        )
        result = result.encode("utf-8").decode("unicode_escape").encode("latin-1").decode("utf-8") if "\\x" in result or "\\n" in result or "\\t" in result else result
        return {"ok": True, "result": result}
    except Exception:
        # Fallback: just decode \\uXXXX
        try:
            s = str(data)
            result = re.sub(
                r'\\[Uu]([0-9a-fA-F]{4,8})',
                lambda m: chr(int(m.group(1), 16)),
                s
            )
            return {"ok": True, "result": result}
        except Exception as e:
            return {"ok": False, "error": str(e)}
