from datetime import datetime

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
        if birthdate > reference_date:
            return {"ok": False, "error": "birthdate cannot be in the future"}
        age_years = reference_date.year - birthdate.year
        age_months = reference_date.month - birthdate.month
        age_days = reference_date.day - birthdate.day
        if age_days < 0:
            age_months -= 1
            prev_month = reference_date.month - 1 if reference_date.month > 1 else 12
            prev_year = reference_date.year if reference_date.month > 1 else reference_date.year - 1
            days_in_prev_month = (datetime(prev_year, prev_month + 1, 1) - datetime(prev_year, prev_month, 1)).days
            age_days += days_in_prev_month
        if age_months < 0:
            age_years -= 1
            age_months += 12
        return {"ok": True, "age_years": age_years, "age_months": age_months, "age_days": age_days}
    except Exception as e:
        return {"ok": False, "error": str(e)}
