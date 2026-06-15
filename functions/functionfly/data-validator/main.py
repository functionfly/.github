import re
from typing import Any


def validate_string(value: Any, rule_value: Any, operator: str) -> bool:
    str_value = str(value).lower()
    str_rule = str(rule_value).lower()
    
    if operator == "equals":
        return str_value == str_rule
    elif operator == "not_equals":
        return str_value != str_rule
    elif operator == "contains":
        return str_rule in str_value
    elif operator == "not_contains":
        return str_rule not in str_value
    elif operator == "starts_with":
        return str_value.startswith(str_rule)
    elif operator == "ends_with":
        return str_value.endswith(str_rule)
    elif operator == "regex":
        try:
            return bool(re.search(str_rule, str_value))
        except re.error:
            return False
    else:
        return False


def validate_number(value: Any, rule_value: Any, operator: str) -> bool:
    try:
        num_value = float(value)
    except (ValueError, TypeError):
        return False
    
    try:
        num_rule = float(rule_value)
    except (ValueError, TypeError):
        return False
    
    if operator == "equals":
        return num_value == num_rule
    elif operator == "not_equals":
        return num_value != num_rule
    elif operator == "greater_than":
        return num_value > num_rule
    elif operator == "less_than":
        return num_value < num_rule
    elif operator == "greater_or_equal":
        return num_value >= num_rule
    elif operator == "less_or_equal":
        return num_value <= num_rule
    elif operator == "between":
        if isinstance(rule_value, (list, tuple)) and len(rule_value) >= 2:
            return num_rule <= num_value <= float(rule_value[1])
        return False
    else:
        return False


def validate_boolean(value: Any, rule_value: Any, operator: str) -> bool:
    bool_value = bool(value)
    
    if operator == "is_true":
        return bool_value is True
    elif operator == "is_false":
        return bool_value is False
    elif operator == "equals":
        return bool_value == bool(rule_value)
    else:
        return False


def validate_list(value: Any, rule_value: Any, operator: str) -> bool:
    if not isinstance(value, (list, tuple)):
        return False
    
    if operator == "contains":
        return rule_value in value
    elif operator == "not_contains":
        return rule_value not in value
    elif operator == "contains_all":
        if isinstance(rule_value, (list, tuple)):
            return all(item in value for item in rule_value)
        return False
    elif operator == "length_equals":
        return len(value) == int(rule_value)
    elif operator == "length_greater_than":
        return len(value) > int(rule_value)
    elif operator == "length_less_than":
        return len(value) < int(rule_value)
    else:
        return False


def validate_email(value: Any, rule_value: Any, operator: str) -> bool:
    email_pattern = r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
    is_valid = bool(re.match(email_pattern, str(value)))
    
    if operator == "is_valid":
        return is_valid
    elif operator == "is_invalid":
        return not is_valid
    else:
        return False


def validate_field(data: dict, field: str, rule_type: str, rule_value: Any) -> tuple[bool, str]:
    if "." in field:
        parts = field.split(".")
        current = data
        for part in parts:
            if isinstance(current, dict) and part in current:
                current = current[part]
            else:
                return False, f"Field '{field}' not found"
        value = current
    else:
        if field not in data:
            return False, f"Field '{field}' not found"
        value = data[field]
    
    operators_string = ["equals", "not_equals", "contains", "not_contains", "starts_with", "ends_with", "regex"]
    operators_number = ["equals", "not_equals", "greater_than", "less_than", "greater_or_equal", "less_or_equal", "between"]
    operators_bool = ["is_true", "is_false", "equals"]
    operators_list = ["contains", "not_contains", "contains_all", "length_equals", "length_greater_than", "length_less_than"]
    operators_email = ["is_valid", "is_invalid"]
    
    if rule_type == "required":
        if value is None or (isinstance(value, str) and value.strip() == ""):
            return False, f"Field '{field}' is required"
        return True, ""
    
    elif rule_type == "type_string":
        if not isinstance(value, str):
            return False, f"Field '{field}' must be a string"
        if "operator" in rule_value:
            op = rule_value.get("operator", "equals")
            if op in operators_string:
                return validate_string(value, rule_value.get("value", ""), op)
        return True, ""
    
    elif rule_type == "type_number":
        if not isinstance(value, (int, float)):
            return False, f"Field '{field}' must be a number"
        if "operator" in rule_value:
            op = rule_value.get("operator", "equals")
            if op in operators_number:
                return validate_number(value, rule_value.get("value", 0), op)
        return True, ""
    
    elif rule_type == "type_boolean":
        if not isinstance(value, bool):
            return False, f"Field '{field}' must be a boolean"
        if "operator" in rule_value:
            op = rule_value.get("operator", "equals")
            if op in operators_bool:
                return validate_boolean(value, rule_value.get("value", False), op)
        return True, ""
    
    elif rule_type == "type_list":
        if not isinstance(value, (list, tuple)):
            return False, f"Field '{field}' must be a list"
        if "operator" in rule_value:
            op = rule_value.get("operator", "contains")
            if op in operators_list:
                return validate_list(value, rule_value.get("value"), op)
        return True, ""
    
    elif rule_type == "email":
        if not isinstance(value, str):
            return False, f"Field '{field}' must be a string"
        if "operator" in rule_value:
            op = rule_value.get("operator", "is_valid")
            if op in operators_email:
                return validate_email(value, rule_value.get("value"), op)
        return True, ""
    
    elif rule_type == "min_length":
        if isinstance(value, (str, list, tuple)):
            if len(value) < int(rule_value):
                return False, f"Field '{field}' must have at least {rule_value} items"
        return True, ""
    
    elif rule_type == "max_length":
        if isinstance(value, (str, list, tuple)):
            if len(value) > int(rule_value):
                return False, f"Field '{field}' must have at most {rule_value} items"
        return True, ""
    
    elif rule_type == "pattern":
        if isinstance(value, str):
            if not re.match(str(rule_value), value):
                return False, f"Field '{field}' does not match pattern"
        return True, ""
    
    else:
        return False, f"Unknown rule type: {rule_type}"


def apply_validation_rules(data: Any, rules: list) -> tuple[bool, list, dict]:
    errors = []
    validated_data = {}
    
    if isinstance(data, dict):
        items_to_validate = [data]
    elif isinstance(data, list):
        items_to_validate = data
    else:
        return False, [{"field": "root", "message": "Data must be a dict or list"}], {}
    
    for item in items_to_validate:
        if not isinstance(item, dict):
            errors.append({"field": "item", "message": "Each item must be an object"})
            continue
        
        for rule in rules:
            field = rule.get("field", "")
            rule_type = rule.get("rule_type", "")
            rule_value = rule.get("rule_value")
            
            is_valid, error_msg = validate_field(item, field, rule_type, rule_value)
            
            if not is_valid:
                errors.append({
                    "field": field,
                    "message": error_msg if error_msg else f"Validation failed for '{field}'"
                })
            
            if field and field in item:
                validated_data[field] = item[field]
    
    is_valid = len(errors) == 0
    
    return is_valid, errors, validated_data


def handler(event: dict[str, Any]) -> dict[str, Any]:
    try:
        data = event.get("data")
        rules = event.get("rules", [])
        
        if data is None:
            return {"ok": False, "error": "data is required"}
        
        if not isinstance(rules, list):
            return {"ok": False, "error": "rules must be a list"}
        
        if len(rules) == 0:
            return {"ok": False, "error": "At least one rule is required"}
        
        for i, rule in enumerate(rules):
            if not isinstance(rule, dict):
                return {"ok": False, "error": f"Rule at index {i} must be an object"}
            if "field" not in rule:
                return {"ok": False, "error": f"Rule at index {i} missing required field 'field'"}
            if "rule_type" not in rule:
                return {"ok": False, "error": f"Rule at index {i} missing required field 'rule_type'"}
        
        valid_rule_types = [
            "required", "type_string", "type_number", "type_boolean", "type_list",
            "email", "min_length", "max_length", "pattern"
        ]
        
        for i, rule in enumerate(rules):
            rule_type = rule.get("rule_type", "")
            if rule_type not in valid_rule_types:
                return {"ok": False, "error": f"Rule at index {i} has invalid rule_type: {rule_type}"}
        
        is_valid, errors, validated_data = apply_validation_rules(data, rules)
        
        return {
            "ok": True,
            "is_valid": is_valid,
            "errors": errors,
            "validated_data": validated_data,
            "rules_applied": len(rules)
        }
        
    except Exception as e:
        return {"ok": False, "error": str(e)}
