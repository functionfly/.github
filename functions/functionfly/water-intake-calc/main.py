from typing import Any


ACTIVITY_MULTIPLIERS = {
    "sedentary": 1.0,
    "light": 1.1,
    "moderate": 1.2,
    "active": 1.35,
    "very_active": 1.5,
}

CLIMATE_MULTIPLIERS = {
    "cold": 0.9,
    "temperate": 1.0,
    "hot": 1.2,
}

ML_PER_OZ = 29.5735


def calculate_base_water_intake(weight_kg: float) -> float:
    base_ml = weight_kg * 30
    return base_ml


def adjust_for_activity(base_ml: float, activity_level: str) -> float:
    multiplier = ACTIVITY_MULTIPLIERS.get(activity_level, 1.0)
    return base_ml * multiplier


def adjust_for_climate(base_ml: float, climate: str) -> float:
    multiplier = CLIMATE_MULTIPLIERS.get(climate, 1.0)
    return base_ml * multiplier


def calculate_daily_goal(weight_kg: float, activity_level: str, climate: str) -> dict:
    base_ml = calculate_base_water_intake(weight_kg)
    
    after_activity = adjust_for_activity(base_ml, activity_level)
    final_ml = adjust_for_climate(after_activity, climate)
    
    final_ml = round(final_ml / 50) * 50
    
    final_oz = round(final_ml / ML_PER_OZ, 1)
    
    return {
        "daily_goal_ml": int(final_ml),
        "daily_goal_oz": final_oz,
        "base_ml": int(base_ml)
    }


def generate_intervals(daily_goal_ml: int) -> list[dict]:
    intervals = []
    
    wake_hour = 7
    sleep_hour = 23
    active_hours = sleep_hour - wake_hour
    
    if active_hours <= 0:
        active_hours = 16
    
    glasses = daily_goal_ml / 250
    glasses_per_hour = glasses / active_hours
    
    hour = wake_hour
    while hour < sleep_hour:
        interval_ml = int(round(glasses_per_hour * 250 / 50) * 50)
        
        if hour < 12:
            time_label = f"{hour}:00 AM"
        elif hour == 12:
            time_label = "12:00 PM"
        else:
            time_label = f"{hour-12}:00 PM"
        
        intervals.append({
            "time": time_label,
            "amount_ml": interval_ml,
            "amount_oz": round(interval_ml / ML_PER_OZ, 1)
        })
        
        hour += 1
    
    intervals.append({
        "time": "Before Bed",
        "amount_ml": int(daily_goal_ml * 0.05 / 50) * 50,
        "amount_oz": round(daily_goal_ml * 0.05 / ML_PER_OZ, 1)
    })
    
    return intervals


def generate_tips(activity_level: str, climate: str) -> list:
    tips = []
    
    tips.append("Drink water first thing in the morning to kickstart your metabolism")
    tips.append("Keep a water bottle with you throughout the day")
    
    if activity_level in ["active", "very_active"]:
        tips.append("Consider drinking 500ml of water for every hour of intense exercise")
        tips.append("Monitor urine color - pale yellow indicates good hydration")
    
    if climate == "hot":
        tips.append("Increase intake during hot weather to compensate for sweating")
        tips.append("Set reminders to drink water every 30 minutes when outdoors")
    elif climate == "cold":
        tips.append("You may not feel thirsty in cold weather, but still need to hydrate regularly")
        tips.append("Hot beverages and soups also contribute to daily fluid intake")
    
    tips.append("Eat water-rich foods like cucumbers, watermelon, and oranges")
    tips.append("Avoid excessive caffeine and alcohol as they can dehydrate")
    
    return tips[:6]


def handler(event: dict[str, Any]) -> dict[str, Any]:
    try:
        weight_kg = event.get("weight_kg")
        activity_level = event.get("activity_level", "moderate").lower().strip()
        climate = event.get("climate", "temperate").lower().strip()
        
        if weight_kg is None:
            return {"ok": False, "error": "weight_kg is required"}
        
        try:
            weight_kg = float(weight_kg)
        except (ValueError, TypeError):
            return {"ok": False, "error": "weight_kg must be a number"}
        
        if weight_kg <= 0:
            return {"ok": False, "error": "weight_kg must be positive"}
        
        if weight_kg < 20:
            return {"ok": False, "error": "weight_kg seems too low (minimum 20kg)"}
        
        if weight_kg > 300:
            return {"ok": False, "error": "weight_kg seems too high (maximum 300kg)"}
        
        valid_activity_levels = ["sedentary", "light", "moderate", "active", "very_active"]
        if activity_level not in valid_activity_levels:
            return {"ok": False, "error": f"activity_level must be one of: {', '.join(valid_activity_levels)}"}
        
        valid_climates = ["cold", "temperate", "hot"]
        if climate not in valid_climates:
            return {"ok": False, "error": f"climate must be one of: {', '.join(valid_climates)}"}
        
        goals = calculate_daily_goal(weight_kg, activity_level, climate)
        intervals = generate_intervals(goals["daily_goal_ml"])
        tips = generate_tips(activity_level, climate)
        
        return {
            "ok": True,
            "daily_goal_ml": goals["daily_goal_ml"],
            "daily_goal_oz": goals["daily_goal_oz"],
            "intervals": intervals,
            "tips": tips,
            "weight_kg": weight_kg,
            "activity_level": activity_level,
            "climate": climate
        }
        
    except Exception as e:
        return {"ok": False, "error": str(e)}
