"""Shipping Cost Calculator - Calculate shipping cost."""


def handler(event):
    try:
        weight_lbs = float(event.get("weight_lbs", 0))
        distance_miles = int(event.get("distance_miles", 0))
        carrier = event.get("carrier", "standard").lower()
        service_level = event.get("service_level", "ground").lower()

        if weight_lbs <= 0:
            return {"ok": False, "error": "weight_lbs must be positive"}
        if weight_lbs > 10000:
            return {"ok": False, "error": "weight_lbs exceeds carrier limit"}
        if distance_miles <= 0:
            return {"ok": False, "error": "distance_miles must be positive"}
        if distance_miles > 20000:
            return {"ok": False, "error": "distance_miles exceeds reasonable limit"}
        if carrier not in ["ups", "fedex", "usps", "standard"]:
            return {"ok": False, "error": "carrier must be ups/fedex/usps/standard"}
        if service_level not in ["ground", "express", "overnight"]:
            return {"ok": False, "error": "service_level must be ground/express/overnight"}

        base_rates = {
            "ups": {"ground": 0.65, "express": 1.50, "overnight": 3.00},
            "fedex": {"ground": 0.60, "express": 1.45, "overnight": 2.95},
            "usps": {"ground": 0.55, "express": 1.40, "overnight": 2.85},
            "standard": {"ground": 0.58, "express": 1.42, "overnight": 2.90}
        }

        rate_per_lb = base_rates[carrier][service_level]
        cost = weight_lbs * rate_per_lb

        distance_factor = 1.0 + (distance_miles / 1000.0) * 0.05
        cost = cost * distance_factor

        if service_level == "express":
            estimated_days = 2
        elif service_level == "overnight":
            estimated_days = 1
        else:
            estimated_days = min(7, max(1, distance_miles // 500 + 1))

        cost = round(cost, 2)

        return {
            "ok": True,
            "cost": cost,
            "estimated_days": estimated_days,
            "carrier": carrier,
            "service_level": service_level
        }
    except (ValueError, TypeError) as e:
        return {"ok": False, "error": f"Invalid input: {str(e)}"}
    except Exception as e:
        return {"ok": False, "error": f"Internal error: {str(e)}"}
