import json

try:
    import yaml
    YAML_AVAILABLE = True
except ImportError:
    YAML_AVAILABLE = False


def handler(event):
    if isinstance(event, dict):
        data = event.get("data", "")
    else:
        data = ""

    if not data:
        return {"ok": False, "error": "data is required"}

    try:
        if isinstance(data, str):
            parsed = json.loads(data)
        else:
            parsed = data

        if YAML_AVAILABLE:
            result = yaml.dump(parsed, default_flow_style=False, sort_keys=False, allow_unicode=True)
        else:
            result = json.dumps(parsed, indent=2)

        return {"ok": True, "result": result}
    except json.JSONDecodeError as e:
        return {"ok": False, "error": f"invalid JSON: {str(e)}"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
