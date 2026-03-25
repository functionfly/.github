import re


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    if not data:
        return {"ok": False, "error": "data is required (RON string)"}
    try:
        import ron
        result = ron.loads(str(data))
        return {"ok": True, "result": result}
    except ImportError:
        try:
            import json
            cleaned = re.sub(r'//[^\n]*', '', str(data))
            cleaned = re.sub(r'/\*.*?\*/', '', cleaned, flags=re.DOTALL)
            cleaned = re.sub(r'\b([A-Za-z_][A-Za-z0-9_]*)\s*\(', r'{"__type__": "\1", "__args__": [', cleaned)
            cleaned = re.sub(r'\)', '}', cleaned)
            result = json.loads(cleaned)
            return {"ok": True, "result": result, "note": "ron library not installed; used simplified parser"}
        except Exception:
            return {"ok": False, "error": "RON parsing requires the 'ron' library. Install with: pip install ron-parser"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
