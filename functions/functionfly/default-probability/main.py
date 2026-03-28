import math

def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be a dict"}
    credit_spread = event.get("credit_spread")
    if credit_spread is None:
        return {"ok": False, "error": "credit_spread is required"}
    try:
        spread = float(credit_spread)
        recovery = float(event.get("recovery_rate", 0.4))
        years = float(event.get("years", 1))
        if recovery >= 1:
            return {"ok": False, "error": "recovery_rate must be less than 1"}
        # Hazard rate approximation: lambda = spread / (1 - recovery)
        hazard_rate = spread / (1 - recovery)
        # Probability of default over horizon
        prob_default = 1 - math.exp(-hazard_rate * years)
        return {
            "ok": True,
            "result": round(prob_default, 8),
            "hazard_rate": round(hazard_rate, 8),
            "survival_probability": round(1 - prob_default, 8)
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
