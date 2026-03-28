from datetime import datetime

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

def handler(event):
    try:
        date_str = event.get("date", "") if isinstance(event, dict) else ""
        country = event.get("country", "") if isinstance(event, dict) else ""
        if not date_str:
            return {"ok": False, "error": "date is required"}
        if not country:
            return {"ok": False, "error": "country is required"}
        try:
            dt = datetime.fromisoformat(date_str.replace("Z", "+00:00"))
        except ValueError:
            return {"ok": False, "error": "invalid date format"}
        month_day = (dt.month, dt.day)
        if country.upper() == "US":
            holidays = US_HOLIDAYS
        elif country.upper() == "UK":
            holidays = UK_HOLIDAYS
        else:
            return {"ok": False, "error": f"unsupported country: {country}"}
        if month_day in holidays:
            return {"ok": True, "is_holiday": True, "holiday_name": holidays[month_day]}
        else:
            return {"ok": True, "is_holiday": False}
    except Exception as e:
        return {"ok": False, "error": str(e)}
