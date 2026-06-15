from typing import Any


def calculate_abandonment_probability(cart_value: float, checkout_steps_completed: int, total_checkout_steps: int, time_on_page_seconds: int) -> float:
    base_probability = 0.0
    
    if total_checkout_steps > 0:
        progress_ratio = checkout_steps_completed / total_checkout_steps
    else:
        progress_ratio = 0
    
    if progress_ratio < 0.3:
        base_probability = 75.0
    elif progress_ratio < 0.5:
        base_probability = 60.0
    elif progress_ratio < 0.75:
        base_probability = 40.0
    else:
        base_probability = 20.0
    
    if cart_value > 200:
        base_probability += 10.0
    elif cart_value > 100:
        base_probability += 5.0
    elif cart_value > 50:
        base_probability += 2.0
    
    if time_on_page_seconds > 600:
        base_probability += 15.0
    elif time_on_page_seconds > 300:
        base_probability += 8.0
    elif time_on_page_seconds < 30:
        base_probability += 5.0
    
    if checkout_steps_completed == 0 and time_on_page_seconds > 120:
        base_probability += 20.0
    
    return min(99.0, max(1.0, round(base_probability, 1)))


def identify_risk_factors(cart_value: float, checkout_steps_completed: int, total_checkout_steps: int, time_on_page_seconds: int) -> list:
    risk_factors = []
    
    if total_checkout_steps > 0:
        progress_ratio = checkout_steps_completed / total_checkout_steps
        if progress_ratio < 0.5:
            risk_factors.append("Low checkout progress - customer may be hesitating at early stages")
    
    if cart_value > 200:
        risk_factors.append(f"High cart value (${cart_value:.2f}) increases purchase hesitation")
    
    if cart_value > 100:
        risk_factors.append("Consider offering financing options or payment plans")
    
    if time_on_page_seconds > 600:
        risk_factors.append("Extended time on checkout page may indicate form complexity or hesitation")
    
    if time_on_page_seconds < 30 and checkout_steps_completed == 0:
        risk_factors.append("Quick abandonment may indicate unexpected costs or UX issues")
    
    if checkout_steps_completed < total_checkout_steps and time_on_page_seconds > 180:
        risk_factors.append("Customer paused during checkout - possible friction point")
    
    if checkout_steps_completed > 2 and checkout_steps_completed < total_checkout_steps:
        risk_factors.append("Customer lost interest after providing information - consider saving progress")
    
    return risk_factors[:4]


def generate_recovery_suggestions(abandonment_probability: float, risk_factors: list, cart_value: float) -> list:
    suggestions = []
    
    if abandonment_probability > 70:
        suggestions.append("Send immediate abandoned cart email with incentive")
        suggestions.append("Display exit-intent popup with special offer")
    
    if cart_value > 100:
        suggestions.append("Offer live chat or chatbot assistance to answer questions")
        suggestions.append("Consider showing security badges and accepted payment methods")
    
    if abandonment_probability > 50:
        suggestions.append("Send follow-up email within 1 hour with cart recovery link")
        suggestions.append("Show progress save option for returning later")
    
    suggestions.append("Simplify checkout by reducing form fields")
    suggestions.append("Add trust signals near payment section")
    
    if not suggestions:
        suggestions.append("Continue monitoring for other abandonment patterns")
    
    return suggestions[:4]


def determine_recommended_action(abandonment_probability: float, cart_value: float, checkout_steps_completed: int) -> str:
    if abandonment_probability >= 80:
        if cart_value > 100:
            return "urgent_recovery_email"
        return "immediate_popup"
    
    if abandonment_probability >= 60:
        if checkout_steps_completed > 2:
            return "save_progress_and_follow_up"
        return "recovery_email_2_hours"
    
    if abandonment_probability >= 40:
        return "follow_up_email_24_hours"
    
    return "no_action_required"


def handler(event: dict[str, Any]) -> dict[str, Any]:
    try:
        cart_value = event.get("cart_value", 0)
        checkout_steps_completed = event.get("checkout_steps_completed", 0)
        total_checkout_steps = event.get("total_checkout_steps", 5)
        time_on_page_seconds = event.get("time_on_page_seconds", 0)
        
        try:
            cart_value = float(cart_value)
        except (ValueError, TypeError):
            return {"ok": False, "error": "cart_value must be a number"}
        
        if cart_value < 0:
            return {"ok": False, "error": "cart_value cannot be negative"}
        
        try:
            checkout_steps_completed = int(checkout_steps_completed)
        except (ValueError, TypeError):
            return {"ok": False, "error": "checkout_steps_completed must be an integer"}
        
        try:
            total_checkout_steps = int(total_checkout_steps)
        except (ValueError, TypeError):
            return {"ok": False, "error": "total_checkout_steps must be an integer"}
        
        try:
            time_on_page_seconds = int(time_on_page_seconds)
        except (ValueError, TypeError):
            return {"ok": False, "error": "time_on_page_seconds must be an integer"}
        
        if time_on_page_seconds < 0:
            return {"ok": False, "error": "time_on_page_seconds cannot be negative"}
        
        if checkout_steps_completed < 0:
            return {"ok": False, "error": "checkout_steps_completed cannot be negative"}
        
        if total_checkout_steps <= 0:
            return {"ok": False, "error": "total_checkout_steps must be positive"}
        
        abandonment_probability = calculate_abandonment_probability(
            cart_value, checkout_steps_completed, total_checkout_steps, time_on_page_seconds
        )
        
        risk_factors = identify_risk_factors(
            cart_value, checkout_steps_completed, total_checkout_steps, time_on_page_seconds
        )
        
        recovery_suggestions = generate_recovery_suggestions(abandonment_probability, risk_factors, cart_value)
        
        recommended_action = determine_recommended_action(abandonment_probability, cart_value, checkout_steps_completed)
        
        return {
            "ok": True,
            "abandonment_probability": abandonment_probability,
            "risk_factors": risk_factors,
            "recovery_suggestions": recovery_suggestions,
            "recommended_action": recommended_action,
            "cart_value": cart_value,
            "checkout_progress": f"{checkout_steps_completed}/{total_checkout_steps}"
        }
        
    except Exception as e:
        return {"ok": False, "error": str(e)}
