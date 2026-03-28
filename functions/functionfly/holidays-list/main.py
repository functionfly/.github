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

def get_holidays(country: str, year: int) -> list:
    """Get holidays for a country and year"""
    holidays = []
    if country.upper() == "US":
        for (month, day), name in US_HOLIDAYS.items():
            holidays.append({"date": f"{year}-{month:02d}-{day:02d}", "name": name})
    elif country.upper() == "UK":
        for (month, day), name in UK_HOLIDAYS.items():
            holidays.append({"date": f"{year}-{month:02d}-{day:02d}", "name": name})
    else:
        return []
    return sorted(holidays, key=lambda x: x["date"])

def handler(event):
    try:
        country = event.get("country", "") if isinstance(event, dict) else ""
        year = event.get("year") if isinstance(event, dict) else None
        if not country:
            return {"ok": False, "error": "country is required"}
        if year is None:
            year = datetime.now().year
        holidays = get_holidays(country, year)
        return {"ok": True, "holidays": holidays}
    except Exception as e:
        return {"ok": False, "error": str(e)}
