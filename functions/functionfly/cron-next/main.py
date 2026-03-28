import re
from datetime import datetime, timedelta
from typing import List, Optional

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
        current += timedelta(minutes=1)
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
        from_time_str = event.get("from_time")
        count = min(event.get("count", 5), 10)
        if not expression:
            return {"ok": False, "error": "expression is required"}
        if not isinstance(expression, str):
            return {"ok": False, "error": "expression must be a string"}
        if not validate_cron_expression(expression):
            return {"ok": False, "error": "invalid cron expression"}
        from_time = None
        if from_time_str:
            try:
                from_time = datetime.fromisoformat(from_time_str.replace("Z", "+00:00"))
            except ValueError:
                return {"ok": False, "error": "invalid from_time format"}
        next_runs = get_next_runs(expression, count, from_time)
        return {"ok": True, "next_runs": next_runs}
    except Exception as e:
        return {"ok": False, "error": f"parsing error: {str(e)}"}
