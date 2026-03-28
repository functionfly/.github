import yaml


def handler(event):
    """Validate YAML syntax."""
    try:
        data = event.get("data")
        if data is None:
            return {"ok": False, "error": "data is required"}

        try:
            parsed = yaml.safe_load(str(data))
            return {"ok": True, "valid": True, "parsed": parsed, "message": "Valid YAML"}
        except yaml.YAMLError as e:
            return {"ok": True, "valid": False, "message": str(e), "error": str(e)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
