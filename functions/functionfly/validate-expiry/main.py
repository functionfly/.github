import datetime


def handler(event):
    expiry = event.get("expiry") if isinstance(event, dict) else None
    if not expiry:
        return {"ok": False, "error": "expiry is required (MM/YY or MM/YYYY)"}
    try:
        s = str(expiry).strip()
        parts = s.replace("-", "/").split("/")
        if len(parts) != 2:
            return {"ok": False, "error": "expiry must be MM/YY or MM/YYYY"}
        month, year = int(parts[0]), int(parts[1])
        if year < 100:
            year += 2000
        if not 1 <= month <= 12:
            return {"ok": True, "result": False, "valid": False, "reason": "invalid month"}
        now = datetime.date.today()
        exp_date = datetime.date(year, month, 1)
        # Card is valid through end of expiry month
        if exp_date.month == 12:
            exp_end = datetime.date(exp_date.year + 1, 1, 1)
        else:
            exp_end = datetime.date(exp_date.year, exp_date.month + 1, 1)
        valid = now < exp_end
        return {
            "ok": True,
            "result": valid,
            "valid": valid,
            "expired": not valid,
            "month": month,
            "year": year,
            "expiry_formatted": f"{month:02d}/{year}"
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
