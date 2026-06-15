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

    if not isinstance(data, str):
        return {"ok": False, "error": "data must be a string"}

    try:
        if YAML_AVAILABLE:
            parsed = yaml.safe_load(data)
        else:
            return {"ok": False, "error": "PyYAML library not available"}

        if parsed is None:
            parsed = {}

        result = json.dumps(parsed, indent=2)
        return {"ok": True, "result": result}
    except yaml.YAMLError as e:
        return {"ok": False, "error": f"invalid YAML: {str(e)}"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
