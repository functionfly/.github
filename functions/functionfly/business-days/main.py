from datetime import datetime, timedelta

US_HOLIDAYS = {
    (1, 1): "New Year's Day",
    (1, 15): "Martin Luther King Jr. Day",
    (2, 19): "Presidents' Day",
    (5, 27): "Memorial Day",
    (6, 19): "Juneteenth",
    (7, 4): "Independence Day",
    (9, 2): "Labor Day",
    (10, 14): "Columbus Day",
    (11, 11): "Veterans Day",
    (11, 28): "Thanksgiving Day",
    (12, 25): "Christmas Day",
}

UK_HOLIDAYS = {
    (1, 1): "New Year's Day",
    (3, 29): "Good Friday",
    (4, 1): "Easter Monday",
    (5, 6): "Early May Bank Holiday",
    (5, 27): "Spring Bank Holiday",
    (8, 26): "Summer Bank Holiday",
    (12, 25): "Christmas Day",
    (12, 26): "Boxing Day",
}

def is_weekend(dt: datetime) -> bool:
    """Check if date is weekend (Saturday or Sunday)"""
    return dt.weekday() >= 5

def is_holiday(dt: datetime, country: str) -> bool:
    """Check if date is a holiday"""
    month_day = (dt.month, dt.day)
    if country.upper() == "US":
        return month_day in US_HOLIDAYS
    elif country.upper() == "UK":
        return month_day in UK_HOLIDAYS
    return False

def handler(event):
    try:
        start_date_str = event.get("start_date", "") if isinstance(event, dict) else ""
        end_date_str = event.get("end_date", "") if isinstance(event, dict) else ""
        exclude_weekends = event.get("exclude_weekends", True) if isinstance(event, dict) else True
        exclude_holidays = event.get("exclude_holidays", False) if isinstance(event, dict) else False
        country = event.get("country", "") if isinstance(event, dict) else ""
        if not start_date_str:
            return {"ok": False, "error": "start_date is required"}
        if not end_date_str:
            return {"ok": False, "error": "end_date is required"}
        try:
            start_date = datetime.fromisoformat(start_date_str.replace("Z", "+00:00"))
            end_date = datetime.fromisoformat(end_date_str.replace("Z", "+00:00"))
        except ValueError:
            return {"ok": False, "error": "invalid date format"}
        if start_date > end_date:
            return {"ok": False, "error": "start_date must be before end_date"}
        total_days = (end_date - start_date).days
        business_days = 0
        current = start_date
        while current <= end_date:
            if exclude_weekends and is_weekend(current):
                current += timedelta(days=1)
                continue
            if exclude_holidays and country and is_holiday(current, country):
                current += timedelta(days=1)
                continue
            business_days += 1
            current += timedelta(days=1)
        return {"ok": True, "business_days": business_days, "total_days": total_days}
    except Exception as e:
        return {"ok": False, "error": str(e)}
