from datetime import datetime, timezone


def handler(event):
    if isinstance(event, dict):
        birth_str = event.get("birthdate", event.get("date", ""))
        as_of_str = event.get("as_of")
    else:
        birth_str, as_of_str = "", None
    if not birth_str:
        return {"ok": False, "error": "Input 'birthdate' is required"}
    try:
        birth = datetime.fromisoformat(birth_str.replace("Z", "+00:00")).date()
        if as_of_str:
            as_of = datetime.fromisoformat(as_of_str.replace("Z", "+00:00")).date()
        else:
            as_of = datetime.now(timezone.utc).date()
        age = as_of.year - birth.year
        if (as_of.month, as_of.day) < (birth.month, birth.day):
            age -= 1
        return {"ok": True, "age": age}
    except Exception as e:
        return {"ok": False, "error": str(e)}
