from datetime import datetime, timedelta

def handler(event):
    try:
        birthdate_str = event.get("birthdate", "") if isinstance(event, dict) else ""
        reference_date_str = event.get("reference_date") if isinstance(event, dict) else None
        if not birthdate_str:
            return {"ok": False, "error": "birthdate is required"}
        try:
            birthdate = datetime.fromisoformat(birthdate_str.replace("Z", "+00:00"))
        except ValueError:
            return {"ok": False, "error": "invalid birthdate format"}
        if reference_date_str:
            try:
                reference_date = datetime.fromisoformat(reference_date_str.replace("Z", "+00:00"))
            except ValueError:
                return {"ok": False, "error": "invalid reference_date format"}
        else:
            reference_date = datetime.now()
        current_year = reference_date.year
        next_birthday = birthdate.replace(year=current_year)
        if next_birthday < reference_date:
            next_birthday = birthdate.replace(year=current_year + 1)
        days_until = (next_birthday - reference_date).days
        return {"ok": True, "days_until_next": days_until, "next_birthday": next_birthday.strftime("%Y-%m-%d")}
    except Exception as e:
        return {"ok": False, "error": str(e)}
