import json


def validate_schema(data, schema):
    """Simple JSON Schema validation (type checking only)."""
    errors = []
    schema_type = schema.get("type")
    if schema_type:
        type_map = {
            "string": str,
            "number": (int, float),
            "integer": int,
            "boolean": bool,
            "array": list,
            "object": dict,
            "null": type(None),
        }
        expected = type_map.get(schema_type)
        if expected and not isinstance(data, expected):
            errors.append(f"Expected type {schema_type}, got {type(data).__name__}")

    if schema_type == "object" and isinstance(data, dict):
        required = schema.get("required", [])
        for field in required:
            if field not in data:
                errors.append(f"Missing required field: {field}")
        props = schema.get("properties", {})
        for key, sub_schema in props.items():
            if key in data:
                sub_errors = validate_schema(data[key], sub_schema)
                errors.extend([f"{key}: {e}" for e in sub_errors])

    return errors


def handler(event):
    """Validate JSON syntax and optionally against a schema."""
    try:
        data = event.get("data")
        if data is None:
            return {"ok": False, "error": "data is required"}

        try:
            parsed = json.loads(str(data))
        except json.JSONDecodeError as e:
            return {"ok": True, "valid": False, "message": str(e), "errors": [str(e)]}

        schema = event.get("schema")
        if schema:
            errors = validate_schema(parsed, schema)
            if errors:
                return {"ok": True, "valid": False, "parsed": parsed, "errors": errors, "message": f"{len(errors)} schema error(s)"}

        return {"ok": True, "valid": True, "parsed": parsed, "errors": [], "message": "Valid JSON"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
