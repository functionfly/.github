import re


def _cson_to_dict(text):
    """Simple CSON parser: handles flat key: value and basic nested objects."""
    def strip_comments(t):
        return re.sub(r'#[^\n]*', '', t)
    
    text = strip_comments(text)
    try:
        import json
        return json.loads(text)
    except Exception:
        pass
    result = {}
    lines = text.strip().splitlines()
    current = result
    stack = [result]
    indent_stack = [0]
    
    for line in lines:
        stripped = line.strip()
        if not stripped:
            continue
        indent = len(line) - len(line.lstrip())
        if ":" in stripped:
            key, _, value = stripped.partition(":")
            key = key.strip().strip("'\"")
            value = value.strip()
            if value == "":
                new_dict = {}
                current[key] = new_dict
                stack.append(current)
                current = new_dict
                indent_stack.append(indent)
            else:
                for q in ['"', "'"]:
                    if value.startswith(q) and value.endswith(q) and len(value) > 1:
                        value = value[1:-1]
                        break
                else:
                    if value.lower() == "true": value = True
                    elif value.lower() == "false": value = False
                    elif value.lower() == "null": value = None
                    else:
                        try: value = int(value)
                        except ValueError:
                            try: value = float(value)
                            except ValueError: pass
                current[key] = value
    return result


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    if not data:
        return {"ok": False, "error": "data is required (CSON string)"}
    try:
        result = _cson_to_dict(str(data))
        return {"ok": True, "result": result}
    except Exception as e:
        return {"ok": False, "error": str(e)}
