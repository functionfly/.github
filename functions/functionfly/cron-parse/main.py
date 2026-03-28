import re
from datetime import datetime, timedelta
from typing import List, Optional

# Cron field names and their ranges
CRON_FIELDS = {
    'minute': (0, 59),
    'hour': (0, 23),
    'day': (1, 31),
    'month': (1, 12),
    'weekday': (0, 6)  # 0=Sunday, 6=Saturday
}

MONTH_NAMES = ['', 'January', 'February', 'March', 'April', 'May', 'June',
               'July', 'August', 'September', 'October', 'November', 'December']

WEEKDAY_NAMES = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday']

def parse_cron_field(field: str, min_val: int, max_val: int) -> List[int]:
    """Parse a single cron field into a list of valid values"""
    if field == '*':
        return list(range(min_val, max_val + 1))

    values = []
    parts = field.split(',')

    for part in parts:
        part = part.strip()
        if '/' in part:
            # Step values like */5 or 10/5
            base, step = part.split('/')
            step = int(step)
            if base == '*':
                values.extend(range(min_val, max_val + 1, step))
            else:
                start = int(base)
                values.extend(range(start, max_val + 1, step))
        elif '-' in part:
            # Ranges like 1-5
            start, end = map(int, part.split('-'))
            values.extend(range(start, end + 1))
        else:
            # Single values
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
        # Parse each field
        minute_vals = parse_cron_field(parts[0], 0, 59)
        hour_vals = parse_cron_field(parts[1], 0, 23)
        day_vals = parse_cron_field(parts[2], 1, 31)
        month_vals = parse_cron_field(parts[3], 1, 12)
        weekday_vals = parse_cron_field(parts[4], 0, 6)

        # Check if all values are in valid ranges
        return (all(0 <= m <= 59 for m in minute_vals) and
                all(0 <= h <= 23 for h in hour_vals) and
                all(1 <= d <= 31 for d in day_vals) and
                all(1 <= m <= 12 for m in month_vals) and
                all(0 <= w <= 6 for w in weekday_vals))

    except (ValueError, IndexError):
        return False

def describe_cron_expression(expression: str) -> str:
    """Generate human-readable description of cron expression"""
    parts = expression.split()

    descriptions = []

    # Minutes
    if parts[0] == '0':
        descriptions.append("At the start of the hour")
    elif parts[0] == '*':
        descriptions.append("every minute")
    else:
        descriptions.append(f"at minute {parts[0]}")

    # Hours
    if parts[1] == '*':
        descriptions.append("every hour")
    elif '-' in parts[1]:
        start, end = parts[1].split('-')
        descriptions.append(f"between {start}:00 and {end}:00")
    else:
        hour = int(parts[1])
        descriptions.append(f"at {hour:02d}:00")

    # Days
    if parts[2] != '*':
        descriptions.append(f"on day {parts[2]} of the month")

    # Months
    if parts[3] != '*':
        if '-' in parts[3]:
            start, end = map(int, parts[3].split('-'))
            descriptions.append(f"from {MONTH_NAMES[start]} to {MONTH_NAMES[end]}")
        else:
            month = int(parts[3])
            descriptions.append(f"in {MONTH_NAMES[month]}")

    # Weekdays
    if parts[4] != '*':
        if '-' in parts[4]:
            start, end = map(int, parts[4].split('-'))
            descriptions.append(f"on {WEEKDAY_NAMES[start]} through {WEEKDAY_NAMES[end]}")
        else:
            weekday = int(parts[4])
            descriptions.append(f"on {WEEKDAY_NAMES[weekday]}")

    # Combine descriptions
    if not descriptions:
        return "Every minute"

    result = descriptions[0]
    for desc in descriptions[1:]:
        if desc.startswith(('at', 'on', 'in', 'between')):
            result += f", {desc}"
        else:
            result += f" {desc}"

    return result.capitalize()

def get_next_runs(expression: str, count: int = 5, from_time: Optional[datetime] = None) -> List[str]:
    """Calculate next run times for cron expression"""
    if from_time is None:
        from_time = datetime.now()

    parts = expression.split()
    minute_vals = parse_cron_field(parts[0], 0, 59)
    hour_vals = parse_cron_field(parts[1], 0, 23)
    day_vals = parse_cron_field(parts[2], 1, 31)
    month_vals = parse_cron_field(parts[3], 1, 12)
    weekday_vals = parse_cron_field(parts[4], 0, 6)

    next_runs = []
    current = from_time.replace(second=0, microsecond=0)

    while len(next_runs) < count:
        # Try next minute
        current += timedelta(minutes=1)

        # Check if this time matches the cron expression
        if (current.minute in minute_vals and
            current.hour in hour_vals and
            current.day in day_vals and
            current.month in month_vals and
            current.weekday() in weekday_vals):

            next_runs.append(current.isoformat() + 'Z')

    return next_runs

def handler(event):
    try:
        expression = event.get("expression")
        get_next_runs_flag = event.get("get_next_runs", True)
        count = min(event.get("count", 5), 10)  # Max 10 runs

        if not expression:
            return {"ok": False, "error": "expression is required"}

        if not isinstance(expression, str):
            return {"ok": False, "error": "expression must be a string"}

        is_valid = validate_cron_expression(expression)

        result = {
            "expression": expression,
            "is_valid": is_valid
        }

        if is_valid:
            result["description"] = describe_cron_expression(expression)

            if get_next_runs_flag:
                result["next_runs"] = get_next_runs(expression, count)

        return {"ok": True, "result": result}

    except Exception as e:
        return {"ok": False, "error": f"parsing error: {str(e)}"}