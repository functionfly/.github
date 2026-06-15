"""Benefits Calculator - Calculate employee benefits value."""


def handler(event):
    try:
        base_salary = float(event.get("base_salary", 0))
        include_retirement = bool(event.get("include_retirement", False))
        include_health = bool(event.get("include_health", False))
        include_pto = bool(event.get("include_pto", False))
        pto_days = int(event.get("pto_days", 0))

        if base_salary <= 0:
            return {"ok": False, "error": "base_salary must be positive"}
        if pto_days < 0 or pto_days > 365:
            return {"ok": False, "error": "pto_days must be between 0 and 365"}

        retirement_value = 0.0
        if include_retirement:
            retirement_value = round(base_salary * 0.06, 2)

        health_value = 0.0
        if include_health:
            health_value = 8000.0

        pto_value = 0.0
        if include_pto and pto_days > 0:
            daily_rate = base_salary / 260.0
            pto_value = round(daily_rate * pto_days, 2)

        total_benefits = round(retirement_value + health_value + pto_value, 2)
        total_compensation = round(base_salary + total_benefits, 2)
        benefits_percent = round((total_benefits / total_compensation) * 100, 2) if total_compensation > 0 else 0

        return {
            "ok": True,
            "retirement_value": retirement_value,
            "health_value": health_value,
            "pto_value": pto_value,
            "total_compensation": total_compensation,
            "benefits_percent": benefits_percent
        }
    except (ValueError, TypeError) as e:
        return {"ok": False, "error": f"Invalid input: {str(e)}"}
    except Exception as e:
        return {"ok": False, "error": f"Internal error: {str(e)}"}
