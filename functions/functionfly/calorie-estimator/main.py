"""Calorie Estimator - Estimate daily calorie needs."""


def handler(event):
    try:
        age = int(event.get("age", 0))
        gender = event.get("gender", "").lower()
        weight_kg = float(event.get("weight_kg", 0))
        height_cm = float(event.get("height_cm", 0))
        activity_level = event.get("activity_level", "sedentary").lower()
        goal = event.get("goal", "maintain").lower()

        if age <= 0 or age > 150:
            return {"ok": False, "error": "age must be between 1 and 150"}
        if gender not in ["male", "female"]:
            return {"ok": False, "error": "gender must be 'male' or 'female'"}
        if weight_kg <= 0:
            return {"ok": False, "error": "weight_kg must be positive"}
        if height_cm <= 0:
            return {"ok": False, "error": "height_cm must be positive"}
        if activity_level not in ["sedentary", "light", "moderate", "active", "very_active"]:
            return {"ok": False, "error": "activity_level must be sedentary/light/moderate/active/very_active"}
        if goal not in ["lose", "maintain", "gain"]:
            return {"ok": False, "error": "goal must be lose/maintain/gain"}

        if gender == "male":
            bmr = 88.362 + (13.397 * weight_kg) + (4.799 * height_cm) - (5.677 * age)
        else:
            bmr = 447.593 + (9.247 * weight_kg) + (3.098 * height_cm) - (4.330 * age)

        activity_multipliers = {
            "sedentary": 1.2,
            "light": 1.375,
            "moderate": 1.55,
            "active": 1.725,
            "very_active": 1.9
        }
        tdee = bmr * activity_multipliers[activity_level]

        goal_adjustments = {"lose": -500, "maintain": 0, "gain": 500}
        daily_calories = round(tdee + goal_adjustments[goal], 2)

        return {
            "ok": True,
            "bmr": round(bmr, 2),
            "tdee": round(tdee, 2),
            "daily_calories": daily_calories
        }
    except (ValueError, TypeError) as e:
        return {"ok": False, "error": f"Invalid input: {str(e)}"}
    except Exception as e:
        return {"ok": False, "error": f"Internal error: {str(e)}"}
