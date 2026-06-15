"""BMI Calculator - Calculate Body Mass Index."""
import math


def handler(event):
    try:
        weight_kg = float(event.get("weight_kg", 0))
        height_cm = float(event.get("height_cm", 0))

        if weight_kg <= 0:
            return {"ok": False, "error": "weight_kg must be positive"}
        if height_cm <= 0:
            return {"ok": False, "error": "height_cm must be positive"}
        if weight_kg > 1000:
            return {"ok": False, "error": "weight_kg exceeds reasonable limit"}
        if height_cm > 300:
            return {"ok": False, "error": "height_cm exceeds reasonable limit"}

        height_m = height_cm / 100.0
        bmi = weight_kg / (height_m * height_m)
        bmi = round(bmi, 2)

        if bmi < 18.5:
            category = "underweight"
        elif bmi < 25:
            category = "normal"
        elif bmi < 30:
            category = "overweight"
        else:
            category = "obese"

        min_weight = round(18.5 * height_m * height_m, 2)
        max_weight = round(24.9 * height_m * height_m, 2)

        return {
            "ok": True,
            "bmi": bmi,
            "category": category,
            "ideal_weight_range": {"min": min_weight, "max": max_weight}
        }
    except (ValueError, TypeError) as e:
        return {"ok": False, "error": f"Invalid input: {str(e)}"}
    except Exception as e:
        return {"ok": False, "error": f"Internal error: {str(e)}"}
