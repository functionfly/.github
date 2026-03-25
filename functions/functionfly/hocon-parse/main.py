def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    if not data:
        return {"ok": False, "error": "data is required (HOCON string)"}
    try:
        import pyhocon
        config = pyhocon.ConfigFactory.parse_string(str(data))
        result = config.as_plain_ordered_dict()
        return {"ok": True, "result": result}
    except ImportError:
        import re, json
        cleaned = re.sub(r'//[^\n]*', '', str(data))
        cleaned = re.sub(r'#[^\n]*', '', cleaned)
        cleaned = re.sub(r'(\w+)\s*=', r'"\1":', cleaned)
        cleaned = re.sub(r'(\w+)\s*:', r'"\1":', cleaned)
        try:
            result = json.loads("{" + cleaned + "}")
            return {"ok": True, "result": result, "note": "pyhocon not installed; used simplified parser"}
        except Exception:
            return {"ok": False, "error": "pyhocon not installed and fallback parser failed. Install with: pip install pyhocon"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
