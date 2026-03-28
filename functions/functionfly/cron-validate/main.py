import re
from typing import List

def parse_cron_field(field: str, min_val: int, max_val: int) -> List[int]:
    """Parse a single cron field into a list of valid values"""
    if field == '*':
        return list(range(min_val, max_val + 1))
    values = []
    parts = field.split(',')
    for part in parts:
        part = part.strip()
        if '/' in part:
            base, step = part.split('/')
            step = int(step)
            if base == '*':
                values.extend(range(min_val, max_val + 1, step))
            else:
                start = int(base)
                values.extend(range(start, max_val + 1, step))
        elif '-' in part:
            start, end = map(int, part.split('-'))
            values.extend(range(start, end + 1))
        else:
            values.append(int(part))
    return sorted(list(set(values)))

def validate_cron_expression(expression: str) -> bool:
    """Validate cron expression format"""
    if not expression or not isinstance(expression, str):
        return False
    parts = expression.strip().split()
    if len(parts) != 5:
        return False
    try:
        minute_vals = parse_cron_field(parts[0], 0, 59)
        hour_vals = parse_cron_field(parts[1], 0, 23)
        day_vals = parse_cron_field(parts[2], 1, 31)
        month_vals = parse_cron_field(parts[3], 1, 12)
        weekday_vals = parse_cron_field(parts[4], 0, 6)
        return (all(0 <= m <= 59 for m in minute_vals) and
                all(0 <= h <= 23 for h in hour_vals) and
                all(1 <= d <= 31 for d in day_vals) and
                all(1 <= m <= 12 for m in month_vals) and
                all(0 <= w <= 6 for w in weekday_vals))
    except (ValueError, IndexError):
        return False

def handler(event):
    try:
        expression = event.get("expression")
        if not expression:
            return {"ok": False, "error": "expression is required"}
        if not isinstance(expression, str):
            return {"ok": False, "error": "expression must be a string"}
        is_valid = validate_cron_expression(expression)
        return {"ok": True, "is_valid": is_valid}
    except Exception as e:
        return {"ok": False, "error": f"validation error: {str(e)}"}
