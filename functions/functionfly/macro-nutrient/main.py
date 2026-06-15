"""
Macro Nutrient Calculator - Calculate macronutrient requirements based on goals.
"""


def handler(event):
    if isinstance(event, dict):
        weight_kg = event.get("weight_kg")
        height_cm = event.get("height_cm")
        age = event.get("age")
        gender = event.get("gender", "male")
        activity_level = event.get("activity_level", "moderate")
        goal = event.get("goal", "maintain")
        body_fat_percent = event.get("body_fat_percent")
    else:
        weight_kg, height_cm, age, gender, activity_level, goal, body_fat_percent = None, None, None, "male", "moderate", "maintain", None

    # Validate required fields
    if weight_kg is None:
        return {"ok": False, "error": "weight_kg is required"}
    if height_cm is None:
        return {"ok": False, "error": "height_cm is required"}
    if age is None:
        return {"ok": False, "error": "age is required"}

    try:
        weight_kg = float(weight_kg)
        height_cm = float(height_cm)
        age = int(age)
    except (ValueError, TypeError):
        return {"ok": False, "error": "weight_kg, height_cm must be numbers, age must be integer"}

    if weight_kg <= 0 or height_cm <= 0 or age <= 0:
        return {"ok": False, "error": "weight_kg, height_cm, age must be positive values"}

    gender = gender.lower()
    if gender not in ("male", "female"):
        return {"ok": False, "error": "gender must be male or female"}

    activity_level = activity_level.lower()
    activity_multipliers = {
        "sedentary": 1.2,
        "light": 1.375,
        "moderate": 1.55,
        "active": 1.725,
        "very_active": 1.9,
    }
    if activity_level not in activity_multipliers:
        return {"ok": False, "error": "activity_level must be sedentary/light/moderate/active/very_active"}

    goal = goal.lower()
    if goal not in ("lose", "maintain", "gain"):
        return {"ok": False, "error": "goal must be lose/maintain/gain"}

    try:
        # Calculate BMR using Mifflin-St Jeor Equation
        if gender == "male":
            bmr = (10 * weight_kg) + (6.25 * height_cm) - (5 * age) + 5
        else:
            bmr = (10 * weight_kg) + (6.25 * height_cm) - (5 * age) - 161

        # Calculate TDEE
        tdee = bmr * activity_multipliers[activity_level]

        # Adjust for goal
        if goal == "lose":
            daily_calories = tdee * 0.8  # 20% deficit
            calorie_adjustment = -20
        elif goal == "gain":
            daily_calories = tdee * 1.15  # 15% surplus
            calorie_adjustment = 15
        else:
            daily_calories = tdee
            calorie_adjustment = 0

        # Calculate lean body mass
        if body_fat_percent is not None:
            try:
                body_fat_percent = float(body_fat_percent)
                lean_mass = weight_kg * (1 - body_fat_percent / 100)
            except (ValueError, TypeError):
                lean_mass = weight_kg * 0.85 if gender == "male" else weight_kg * 0.75
        else:
            lean_mass = weight_kg * 0.85 if gender == "male" else weight_kg * 0.75

        # Protein: 1.6-2.2g per kg of lean mass for goals
        protein_modifiers = {"lose": 2.2, "maintain": 1.8, "gain": 2.0}
        protein_grams = lean_mass * protein_modifiers[goal]
        protein_calories = protein_grams * 4

        # Fat: 0.7-1g per kg of body weight
        fat_grams = weight_kg * 0.8
        fat_calories = fat_grams * 9

        # Carbs: remaining calories
        carbs_calories = max(0, daily_calories - protein_calories - fat_calories)
        carbs_grams = carbs_calories / 4

        return {
            "ok": True,
            "bmr": round(bmr, 0),
            "tdee": round(tdee, 0),
            "daily_calories": round(daily_calories, 0),
            "macros": {
                "protein_grams": round(protein_grams, 1),
                "protein_percent": round(protein_calories / daily_calories * 100, 1),
                "carbs_grams": round(carbs_grams, 1),
                "carbs_percent": round(carbs_calories / daily_calories * 100, 1),
                "fat_grams": round(fat_grams, 1),
                "fat_percent": round(fat_calories / daily_calories * 100, 1),
            },
            "meal_timing": {
                "meals_per_day": 4,
                "calories_per_meal_approx": round(daily_calories / 4, 0),
                "protein_per_meal_approx": round(protein_grams / 4, 1),
            },
            "recommendations": [
                "Spread protein intake across all meals",
                "Time carbs around workouts for energy",
                "Include healthy fats for hormone health",
                "Stay hydrated (30ml per kg body weight)",
                "Adjust based on progress after 2 weeks",
            ],
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}