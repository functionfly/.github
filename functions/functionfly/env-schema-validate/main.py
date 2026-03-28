import re


def handler(event):
    """Validate environment variables against a schema."""
    try:
        env = event.get("env")
        schema = event.get("schema")
        if env is None or schema is None:
            return {"ok": False, "error": "env and schema are required"}

        errors = []
        warnings = []

        for var_name, var_schema in schema.items():
            value = env.get(var_name)
            required = var_schema.get("required", False)
            var_type = var_schema.get("type")
            pattern = var_schema.get("pattern")
            default = var_schema.get("default")

            if value is None:
                if required:
                    if default is not None:
                        warnings.append(f"{var_name} not set, using default: {default}")
                    else:
                        errors.append(f"{var_name} is required but not set")
                continue

            # Type validation
            if var_type == "number":
                try:
                    float(value)
                except (ValueError, TypeError):
                    errors.append(f"{var_name} must be a number, got: {value}")
            elif var_type == "boolean":
                if str(value).lower() not in ("true", "false", "1", "0", "yes", "no"):
                    errors.append(f"{var_name} must be a boolean, got: {value}")
            elif var_type == "url":
                if not str(value).startswith(("http://", "https://")):
                    errors.append(f"{var_name} must be a valid URL, got: {value}")

            # Pattern validation
            if pattern:
                if not re.match(pattern, str(value)):
                    errors.append(f"{var_name} does not match pattern {pattern}")

        return {"ok": True, "valid": len(errors) == 0, "errors": errors, "warnings": warnings}
    except Exception as e:
        return {"ok": False, "error": str(e)}
