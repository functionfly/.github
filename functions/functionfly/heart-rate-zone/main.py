"""Heart Rate Zone Calculator - Calculate heart rate training zones."""


def handler(event):
    try:
        age = int(event.get("age", 0))
        resting_hr = int(event.get("resting_hr", 70))
        max_hr_input = event.get("max_hr")

        if age <= 0 or age > 120:
            return {"ok": False, "error": "age must be between 1 and 120"}
        if resting_hr < 40 or resting_hr > 120:
            return {"ok": False, "error": "resting_hr must be between 40 and 120"}

        if max_hr_input is not None:
            max_hr = int(max_hr_input)
            if max_hr < 100 or max_hr > 250:
                return {"ok": False, "error": "max_hr must be between 100 and 250"}
        else:
            max_hr = 220 - age

        if max_hr <= resting_hr:
            return {"ok": False, "error": "max_hr must be greater than resting_hr"}

        def zone_bounds(lower_pct, upper_pct):
            lower = round(max_hr * lower_pct)
            upper = round(max_hr * upper_pct)
            return {"lower": lower, "upper": upper}

        zones = {
            "zone1": zone_bounds(0.50, 0.60),
            "zone2": zone_bounds(0.60, 0.70),
            "zone3": zone_bounds(0.70, 0.80),
            "zone4": zone_bounds(0.80, 0.90),
            "zone5": zone_bounds(0.90, 1.00)
        }

        return {
            "ok": True,
            "max_hr": max_hr,
            "zone1": zones["zone1"],
            "zone2": zones["zone2"],
            "zone3": zones["zone3"],
            "zone4": zones["zone4"],
            "zone5": zones["zone5"]
        }
    except (ValueError, TypeError) as e:
        return {"ok": False, "error": f"Invalid input: {str(e)}"}
    except Exception as e:
        return {"ok": False, "error": f"Internal error: {str(e)}"}
