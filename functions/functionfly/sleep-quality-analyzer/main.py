from datetime import datetime, timedelta
from typing import Any


SLEEP_STAGE_LABELS = {
    "awake": {"label": "Awake", "quality_weight": 0.0, "color": "#ff6b6b"},
    "rem": {"label": "REM Sleep", "quality_weight": 0.25, "color": "#4ecdc4"},
    "light": {"label": "Light Sleep", "quality_weight": 0.35, "color": "#45b7d1"},
    "deep": {"label": "Deep Sleep", "quality_weight": 0.40, "color": "#2c3e50"},
}


def parse_time(time_str: str) -> datetime:
    formats = [
        "%H:%M",
        "%I:%M %p",
        "%H:%M:%S",
        "%I:%M:%S %p",
        "%Y-%m-%d %H:%M",
        "%Y-%m-%d %H:%M:%S",
        "%Y-%m-%dT%H:%M:%S",
        "%Y-%m-%dT%H:%M:%SZ",
        "%Y-%m-%dT%H:%M:%S%z",
    ]
    
    for fmt in formats:
        try:
            return datetime.strptime(time_str, fmt)
        except ValueError:
            continue
    
    raise ValueError(f"Could not parse time: {time_str}")


def calculate_sleep_duration(bed_time: str, wake_time: str) -> float:
    bed = parse_time(bed_time)
    wake = parse_time(wake_time)
    
    if wake < bed:
        wake += timedelta(days=1)
    
    duration = wake - bed
    hours = duration.total_seconds() / 3600
    
    return round(hours, 2)


def calculate_sleep_efficiency(total_hours: float, sleep_stages: list) -> float:
    if not sleep_stages or total_hours <= 0:
        ideal_sleep_hours = min(total_hours, 8.0)
        efficiency = (ideal_sleep_hours / 8.0) * 100
        return round(min(100, efficiency), 1)
    
    actual_sleep = sum(stage.get("duration", 0) for stage in sleep_stages if stage.get("stage", "").lower() != "awake")
    
    time_in_bed = total_hours
    
    if time_in_bed <= 0:
        return 0.0
    
    efficiency = (actual_sleep / time_in_bed) * 100
    
    return round(min(100, efficiency), 1)


def analyze_sleep_stages(sleep_stages: list, total_hours: float) -> dict:
    if not sleep_stages:
        rem_ideal = 0.20 * total_hours
        deep_ideal = 0.20 * total_hours
        light_ideal = 0.50 * total_hours
        awake_time = 0.10 * total_hours
        
        return {
            "rem": {"hours": round(rem_ideal, 2), "percent": 20.0, "status": "estimated"},
            "deep": {"hours": round(deep_ideal, 2), "percent": 20.0, "status": "estimated"},
            "light": {"hours": round(light_ideal, 2), "percent": 50.0, "status": "estimated"},
            "awake": {"hours": round(awake_time, 2), "percent": 10.0, "status": "estimated"},
        }
    
    breakdown = {}
    total_stage_minutes = sum(stage.get("duration", 0) for stage in sleep_stages)
    
    for stage_data in sleep_stages:
        stage_name = stage_data.get("stage", "unknown").lower()
        duration = stage_data.get("duration", 0)
        
        if stage_name in SLEEP_STAGE_LABELS:
            percent = (duration / total_stage_minutes * 100) if total_stage_minutes > 0 else 0
            breakdown[stage_name] = {
                "hours": round(duration / 60, 2),
                "percent": round(percent, 1),
                "status": "recorded",
                "label": SLEEP_STAGE_LABELS[stage_name]["label"]
            }
    
    for stage_name, info in SLEEP_STAGE_LABELS.items():
        if stage_name not in breakdown:
            breakdown[stage_name] = {
                "hours": 0,
                "percent": 0,
                "status": "not_recorded",
                "label": info["label"]
            }
    
    return breakdown


def analyze_heart_rate(heart_rate_data: list) -> dict:
    if not heart_rate_data or len(heart_rate_data) == 0:
        return {"analyzed": False, "message": "No heart rate data provided"}
    
    valid_hr = [hr for hr in heart_rate_data if isinstance(hr, (int, float)) and hr > 0]
    
    if not valid_hr:
        return {"analyzed": False, "message": "No valid heart rate values"}
    
    avg_hr = sum(valid_hr) / len(valid_hr)
    min_hr = min(valid_hr)
    max_hr = max(valid_hr)
    
    hr_category = "normal"
    if avg_hr > 100:
        hr_category = "elevated"
    elif avg_hr < 50:
        hr_category = "low"
    
    return {
        "analyzed": True,
        "average_hr": round(avg_hr, 1),
        "min_hr": min_hr,
        "max_hr": max_hr,
        "category": hr_category,
        "data_points": len(valid_hr)
    }


def calculate_quality_score(total_hours: float, efficiency: float, sleep_stages: list, heart_rate_info: dict) -> int:
    score = 0.0
    
    if 7 <= total_hours <= 9:
        score += 35
    elif 6 <= total_hours < 7 or 9 < total_hours <= 10:
        score += 25
    elif 5 <= total_hours < 6 or 10 < total_hours <= 12:
        score += 15
    else:
        score += 5
    
    score += (efficiency / 100) * 30
    
    if sleep_stages and len(sleep_stages) > 0:
        rem_found = any(stage.get("stage", "").lower() == "rem" for stage in sleep_stages)
        deep_found = any(stage.get("stage", "").lower() == "deep" for stage in sleep_stages)
        
        if rem_found:
            score += 15
        if deep_found:
            score += 15
    else:
        score += 15
    
    if heart_rate_info.get("analyzed"):
        if heart_rate_info.get("category") == "normal":
            score += 5
    
    return int(min(100, max(0, round(score))))


def generate_recommendations(total_hours: float, efficiency: float, quality_score: int, heart_rate_info: dict) -> list:
    recommendations = []
    
    if total_hours < 7:
        recommendations.append("Try to get at least 7-9 hours of sleep per night for optimal health")
    
    if total_hours > 9:
        recommendations.append("Consider checking if oversleeping is due to poor sleep quality")
    
    if efficiency < 85:
        recommendations.append("Improve sleep environment: reduce noise, light, and screen exposure before bed")
        recommendations.append("Avoid caffeine and heavy meals 4-6 hours before bedtime")
    
    if quality_score < 50:
        recommendations.append("Consider establishing a consistent sleep schedule, even on weekends")
        recommendations.append("Limit blue light exposure 1-2 hours before sleep")
    
    if heart_rate_info.get("analyzed"):
        if heart_rate_info.get("category") == "elevated":
            recommendations.append("Your resting heart rate is elevated - consider relaxation techniques before bed")
        elif heart_rate_info.get("category") == "low":
            recommendations.append("Your heart rate is quite low during sleep - consult a doctor if this is unusual for you")
    
    if not recommendations:
        recommendations.append("Your sleep patterns look good! Maintain your current sleep habits")
    
    return recommendations[:5]


def handler(event: dict[str, Any]) -> dict[str, Any]:
    try:
        bed_time = event.get("bed_time", "")
        wake_time = event.get("wake_time", "")
        heart_rate_data = event.get("heart_rate_data", [])
        sleep_stages = event.get("sleep_stages", [])
        
        if not bed_time:
            return {"ok": False, "error": "bed_time is required"}
        
        if not wake_time:
            return {"ok": False, "error": "wake_time is required"}
        
        try:
            total_sleep_hours = calculate_sleep_duration(bed_time, wake_time)
        except ValueError as e:
            return {"ok": False, "error": f"Invalid time format: {str(e)}"}
        
        if total_sleep_hours <= 0 or total_sleep_hours > 24:
            return {"ok": False, "error": "Invalid sleep duration calculated from provided times"}
        
        sleep_efficiency = calculate_sleep_efficiency(total_sleep_hours, sleep_stages)
        
        sleep_stages_breakdown = analyze_sleep_stages(sleep_stages, total_sleep_hours)
        
        heart_rate_analysis = analyze_heart_rate(heart_rate_data)
        
        quality_score = calculate_quality_score(
            total_sleep_hours, 
            sleep_efficiency, 
            sleep_stages, 
            heart_rate_analysis
        )
        
        recommendations = generate_recommendations(
            total_sleep_hours, 
            sleep_efficiency, 
            quality_score,
            heart_rate_analysis
        )
        
        return {
            "ok": True,
            "total_sleep_hours": total_sleep_hours,
            "sleep_efficiency": sleep_efficiency,
            "sleep_stages_breakdown": sleep_stages_breakdown,
            "quality_score": quality_score,
            "recommendations": recommendations,
            "heart_rate_analysis": heart_rate_analysis
        }
        
    except Exception as e:
        return {"ok": False, "error": str(e)}
