import json


def _validate(data, schema, path=""):
    errors = []
    stype = schema.get("type")
    if stype:
        type_map = {"string": str, "number": (int, float), "integer": int, "boolean": bool, "object": dict, "array": list, "null": type(None)}
        expected = type_map.get(stype)
        if expected and not isinstance(data, expected):
            errors.append(f"{path or 'root'}: expected {stype}, got {type(data).__name__}")
            return errors
    if isinstance(data, dict) and "properties" in schema:
        for req in schema.get("required", []):
            if req not in data:
                errors.append(f"{path or 'root'}: missing required field '{req}'")
        for k, sub in schema["properties"].items():
            if k in data:
                errors.extend(_validate(data[k], sub, f"{path}.{k}".lstrip(".")))
    if isinstance(data, list) and "items" in schema:
        for i, item in enumerate(data):
            errors.extend(_validate(item, schema["items"], f"{path}[{i}]"))
    if isinstance(data, str):
        mn = schema.get("minLength"); mx = schema.get("maxLength")
        if mn is not None and len(data) < mn: errors.append(f"{path}: string too short (min {mn})")
        if mx is not None and len(data) > mx: errors.append(f"{path}: string too long (max {mx})")
    if isinstance(data, (int, float)):
        mn = schema.get("minimum"); mx = schema.get("maximum")
        if mn is not None and data < mn: errors.append(f"{path}: value {data} < minimum {mn}")
        if mx is not None and data > mx: errors.append(f"{path}: value {data} > maximum {mx}")
    enum = schema.get("enum")
    if enum is not None and data not in enum:
        errors.append(f"{path}: value not in enum {enum}")
    return errors


def handler(event):
    data = event.get("data") if isinstance(event, dict) else None
    schema = event.get("schema")
    if data is None:
        return {"ok": False, "error": "data is required"}
    if not schema or not isinstance(schema, dict):
        return {"ok": False, "error": "schema must be a JSON Schema object"}
    try:
        try:
            import jsonschema
            try:
                jsonschema.validate(data, schema)
                return {"ok": True, "result": True, "valid": True, "errors": []}
            except jsonschema.ValidationError as e:
                return {"ok": True, "result": False, "valid": False, "errors": [e.message]}
            except jsonschema.SchemaError as e:
                return {"ok": False, "error": f"Invalid schema: {e.message}"}
        except ImportError:
            errors = _validate(data, schema)
            valid = len(errors) == 0
            return {"ok": True, "result": valid, "valid": valid, "errors": errors,
                    "note": "jsonschema not installed; using basic validator"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
